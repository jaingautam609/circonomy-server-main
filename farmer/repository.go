package farmer

import (
	"circonomy-server/dbutil"
	"circonomy-server/models"
	"circonomy-server/repobase"
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"time"
)

type Repository struct {
	repobase.Base
}

func NewRepository(sqlDB *sqlx.DB) *Repository {
	return &Repository{
		repobase.NewBase(sqlDB),
	}
}

func (r *Repository) txx(fn func(txRepo *Repository) error) error {
	return dbutil.WithTransaction(r.DB(), func(tx *sqlx.Tx) error {
		repoCopy := *r
		repoCopy.Base = r.Base.CopyWithTX(tx)
		return fn(&repoCopy)
	})
}

func (r *Repository) updateOTP(otpReq models.SendOTP) error {
	return r.txx(func(txRepo *Repository) error {
		err := txRepo.deletePreviousOTP(otpReq)
		if err != nil {
			return err
		}
		return txRepo.sendOTP(otpReq)
	})
}

func (r *Repository) sendOTP(otpReq models.SendOTP) error {
	args := []interface{}{
		otpReq.PhoneNumber,
		otpReq.OTP,
		otpReq.Type,
		otpReq.Expiry,
		otpReq.CountryCode,
	}
	SQL := `INSERT INTO user_otp(input,otp,type,expiry,country_code) VALUES($1,$2,$3,$4,$5) `
	_, err := r.Exec(SQL, args...)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) verifyOTP(otpReq checkOTPRequest) (bool, error) {
	var isOTPPresent bool

	SQL := `SELECT count(id) > 0
			FROM user_otp
			WHERE otp = $1
  			AND input = $2
			AND country_code = $3
  			AND verified_at IS NULL
  			AND archived_at IS NULL
  			AND expiry > now();`
	err := r.Get(&isOTPPresent, SQL, otpReq.OTP, otpReq.PhoneNumber, otpReq.CountryCode)
	if err != nil {
		return false, err
	}
	return isOTPPresent, nil
}

func (r *Repository) userPhoneExists(phoneNumber string, countryCode string) (string, error) {
	var userID string

	SQL := `SELECT id
			FROM users
			WHERE number = $1
			AND country_code = $2
			AND account_type = $3
  			AND archived_at IS NULL;`
	err := r.Get(&userID, SQL, phoneNumber, countryCode, models.UserAccountTypeFarmer)
	if err != nil {
		return "", err
	}
	return userID, nil
}

// updateAccountDetails add account details name, age, gender, address, profile image id
func (r *Repository) updateAccountDetails(accountDetails updateFarmerDetails, farmerID string) error {
	// arguments for sql query
	args := []interface{}{
		accountDetails.Name,
		accountDetails.Age,
		accountDetails.Gender,
		accountDetails.Address,
		accountDetails.ImageID,
		accountDetails.AadhaarNo,
		accountDetails.AadhaarImageID,
		farmerID,
	}

	SQL := `UPDATE users
				SET name                = $1,
					age                 = $2,
					gender              = $3,
					address             = $4,
					upload_id    		= $5,
					aadhaar_no          = $6,
					aadhaar_no_image_id = $7,
					updated_at          = now()
				WHERE id = $8;`
	_, err := r.Exec(SQL, args...)
	if err != nil {
		return err
	}
	return nil
}

// updateFarmerRegistrationStatus set correct registrations status
func (r *Repository) updateFarmerRegistrationStatus(farmerID string, status registrationStatus) error {
	SQL := `UPDATE farmer_details 
			 SET 
			    registration_status = $2,
			    updated_at = now()
			 WHERE user_id = $1`
	_, err := r.Exec(SQL, farmerID, string(status))
	if err != nil {
		return err
	}
	return nil
}

// addFarm add user farms
func (r *Repository) addFarm(farm addFarmDetails, farmerID string, status registrationStatus, croppingPattern string) (farmID string, err error) {
	location := fmt.Sprintf("(%v,%v)", farm.FarmLatitude, farm.FarmLongitude)
	// arguments for sql query
	args := []interface{}{
		farm.FieldSize,
		farm.FieldSizeUnit,
		croppingPattern,
		location,
		pq.Array(farm.FarmImageIDs),
		farmerID,
		farm.Landmark,
	}

	err = r.txx(func(txRepo *Repository) error {
		// language = SQL
		SQL := `INSERT INTO farms
			(size, size_unit, cropping_pattern, farm_location, farm_images_ids, farmer_id, landmark)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id;`
		err = txRepo.Get(&farmID, SQL, args...)
		if err != nil {
			return err
		}

		SQL = `UPDATE farmer_details
			SET registration_status = $1,
				updated_at          = now()
			WHERE user_id = $2::uuid;`
		_, err = txRepo.Exec(SQL, string(status), farmerID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return farmID, nil
}

// insertFarmer insert farmer in user table
func (r *Repository) insertFarmer(phoneNo string, countryCode string) (string, error) {
	var farmerID string
	err := r.txx(func(txRepo *Repository) error {
		SQL := `INSERT INTO users(number,country_code,account_type) VALUES($1, $2, $3) RETURNING id`
		err := txRepo.Get(&farmerID, SQL, phoneNo, countryCode, models.UserAccountTypeFarmer)
		if err != nil {
			return err
		}

		SQL = `INSERT INTO farmer_details (user_id, registration_status)
			VALUES ($1, $2)`
		_, err = txRepo.Exec(SQL, farmerID, verifyAccount)
		if err != nil {
			return err
		}
		return nil
	})

	return farmerID, err
}

// getCrops get admin crops
func (r *Repository) getCrops(filters models.GenericFilters) ([]crop, error) {
	args := []interface{}{
		filters.SearchString,
		filters.Limit,
		filters.Limit * filters.Page,
		filters.SearchString == "",
	}
	crops := make([]crop, 0)
	SQL := `SELECT id,
				crop_name,
				season,
				crop_image_id
			FROM crops
			WHERE archived_at IS NULL
			 AND (true=$4 OR crop_name ILIKE '%%' || $1 || '%%')
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`
	err := r.Select(&crops, SQL, args...)
	if err != nil {
		return crops, err
	}
	return crops, nil
}

// fetchFarmerDetails fetch farmer details
func (r *Repository) fetchFarmerDetails(farmerID string) (farmerInfo farmerDetails, err error) {

	SQL := `SELECT u.id,
				   u.name as name,
				   u.age as age,
				   u.gender as gender,
				   u.address as address,
				   u.number,
				   u.country_code,
				   u.upload_id as profile_image_id,
				   u.aadhaar_no as aadhaar_no,
				   u.aadhaar_no_image_id as aadhaar_no_image_id,
				   fd.registration_status                    AS registration_status
			FROM users u
			JOIN farmer_details fd ON u.id = fd.user_id
			WHERE u.id = $1
			  AND archived_at IS NULL`

	err = r.Get(&farmerInfo, SQL, farmerID)
	if err != nil {
		return farmerInfo, err
	}
	return farmerInfo, nil
}

// fetchCropsCount fetch crops count
func (r *Repository) fetchCropsCount(filters models.GenericFilters) (int, error) {
	var count int
	SQL := `SELECT COUNT(id)
			 FROM crops
			 WHERE crop_name ILIKE coalesce('%%' || $1 || '%%', crop_name)
			  AND archived_at IS NULL`
	err := r.Get(&count, SQL, filters.SearchString)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// fetchFarmDetails fetch farm details
func (r *Repository) fetchFarmDetails(farmerID string) (farmInfo []farmDetails, err error) {
	SQL := `SELECT id,
       			   landmark,
				   size,
				   size_unit,
				   cropping_pattern,
				   farm_location,
				   farm_images_ids
			FROM farms
			WHERE farmer_id = $1
			  AND archived_at IS NULL`
	err = r.Select(&farmInfo, SQL, farmerID)
	if err != nil {
		return farmInfo, err
	}
	return farmInfo, nil
}

// fetchImagePath fetch image path
func (r *Repository) fetchImagePath(imageIDs pq.StringArray) (images []image, err error) {
	SQL := `SELECT uploads.id,
				   uploads.path,
				   uploads.type,
				   crop_images.crop_status
			FROM uploads
			join crop_images on uploads.id = crop_images.image_id
			WHERE uploads.id = ANY ($1)`
	err = r.Select(&images, SQL, imageIDs)
	if err != nil {
		return images, err
	}
	return images, nil
}

// fetchAdminCropImagePath fetch image path
func (r *Repository) fetchAdminCropImagePath(imageIDs pq.StringArray) (images []image, err error) {
	SQL := `SELECT uploads.id,
				   uploads.path,
				   uploads.type
			FROM uploads
			join crops on uploads.id = crops.crop_image_id
			WHERE uploads.id = ANY ($1)`
	err = r.Select(&images, SQL, imageIDs)
	if err != nil {
		return images, err
	}
	return images, nil
}

// fetchProfileImagePath fetch profileimage path
func (r *Repository) fetchProfileImagePath(imageIDs pq.StringArray) (images []image, err error) {
	SQL := `SELECT uploads.id,
				   uploads.path,
				   uploads.type
			FROM uploads
			WHERE uploads.id = ANY ($1)`
	err = r.Select(&images, SQL, imageIDs)
	if err != nil {
		return images, err
	}
	return images, nil
}

// addPreferredCrops add farmer preferred crops
func (r *Repository) addPreferredCrops(farmerID string, cropIDs []string, status registrationStatus) error {
	return r.txx(func(txRepo *Repository) error {

		qBuilder := squirrel.Insert("farmer_preferred_crop").Columns("farmer_id", "crop_id")
		for idx := range cropIDs {
			qBuilder = qBuilder.Values(farmerID, cropIDs[idx])
		}
		finalSQL, args2, err := qBuilder.PlaceholderFormat(squirrel.Dollar).ToSql()
		_, err = txRepo.Exec(finalSQL, args2...)
		if err != nil {
			return err
		}

		SQL := `UPDATE farmer_details
				SET registration_status = $1,
					updated_at          = now()
				WHERE user_id = $2`
		_, err = txRepo.Exec(SQL, status, farmerID)
		return err
	})
}

// deleteFarm delete farm
func (r *Repository) deleteFarm(farmID, farmerID string) (int64, error) {
	SQL := `UPDATE farms
			SET archived_at = now()
			WHERE id = $1
			  AND farmer_id = $2
			  AND archived_at IS NULL`
	rows, err := r.Exec(SQL, farmID, farmerID)
	if err != nil {
		return 0, err
	}
	rowsAffected, _ := rows.RowsAffected()
	return rowsAffected, nil
}

// markVerifiedOTP mark verified OTP
func (r *Repository) markVerifiedOTP(otpReq checkOTPRequest) error {
	SQL1 := `UPDATE user_otp
			 SET verified_at=now(),
			     archived_at=now()
			 WHERE otp = $1
			  AND input = $2
			  AND country_code =$3
			  AND archived_at IS NULL
			  AND verified_at IS NULL`
	_, err := r.Exec(SQL1, otpReq.OTP, otpReq.PhoneNumber, otpReq.CountryCode)
	if err != nil {
		return errors.Wrap(err, "markVerifiedOTP")
	}
	return nil
}

// deletePreviousOTP delete previous OTP
func (r *Repository) deletePreviousOTP(otpReq models.SendOTP) error {
	SQL := `UPDATE user_otp
			SET archived_at=now()
			WHERE input = $1
			  AND country_code = $2
			AND expiry > now()
			AND archived_at is null
			AND verified_at is null`
	_, err := r.Exec(SQL, otpReq.PhoneNumber, otpReq.CountryCode)
	if err != nil {
		return errors.Wrap(err, "deletePreviousOTP")
	}
	return nil
}

// fetchFarmerFarmIDs fetch farmer farm ids
func (r *Repository) fetchFarmerFarmIDs(farmerID string) ([]string, error) {
	farmIds := make([]string, 0)
	SQL := `SELECT farms.id
			FROM farms
			WHERE farms.farmer_id = $1
			  AND farms.archived_at IS NULL`
	err := r.Select(&farmIds, SQL, farmerID)
	if err != nil {
		return farmIds, errors.Wrap(err, "fetchFarmerFarmIDs")
	}
	return farmIds, nil
}

// fetchFarmCrops fetch farmer crops
func (r *Repository) fetchFarmCrops(farmIDs pq.StringArray, filters models.CropsFilters) ([]farmCrops, error) {
	cropsFiltered := make([]farmCrops, 0)

	status := pq.Array(filters.CropStages)

	args := []interface{}{
		farmIDs, // use check
		filters.SearchString,
		filters.Limit,
		filters.Limit * filters.Page,
		status,
		filters.SearchString == "",
	}

	SQL := `SELECT farm_crops.id,
       			   farm_crops.farm_id,
       			   farms.landmark,
       			   farms.farm_location,
				   crops.crop_name,
				   farm_crops.crop_area,
				   farm_crops.crop_area_unit,
				   array_remove(ARRAY_AGG(distinct uploads.id), NULL) AS farm_crop_images_ids
			FROM farm_crops
					 JOIN crops on 	crops.id = farm_crops.crop_id
			    	 JOIN farms on farm_crops.farm_id = farms.id
					 LEFT JOIN crop_images on crop_images.farmer_crop_id = farm_crops.id AND crop_images.crop_status = any($5)
					 LEFT JOIN uploads on uploads.id = crop_images.image_id
			WHERE farm_crops.farm_id = ANY($1)
			  AND farm_crops.archived_at IS NULL
			  AND (true=$6 OR crops.crop_name ILIKE '%%' || $2 || '%%')
			  AND farm_crops.crop_stage = any($5)
			  AND crop_images.archived_at IS NULL 
			GROUP BY farm_crops.id, crops.crop_name, farm_crops.created_at, farms.id
			ORDER BY farm_crops.created_at DESC
			LIMIT $3 OFFSET $4`

	err := r.Select(&cropsFiltered, SQL, args...)
	if err != nil {
		return cropsFiltered, errors.Wrap(err, "fetchFarmCrops")
	}
	return cropsFiltered, nil
}

// fetchFarmerCropStages fetch Farmer crops stages
func (r *Repository) fetchFarmerCropStages(cropIds pq.StringArray) ([]cropStageTime, error) {
	cropStageTimes := make([]cropStageTime, 0)

	SQL := `SELECT fcs.id,
				   fcs.stage,
				   fcs.starting_time,
				   fcs.farm_crop_id
			FROM farm_crop_stages fcs
				LEFT JOIN farm_crops fc on fcs.farm_crop_id = fc.id
			WHERE fc.id = ANY ($1)`

	err := r.Select(&cropStageTimes, SQL, cropIds)
	if err != nil {
		return nil, errors.Wrap(err, "fetchFarmerCropStages")
	}

	return cropStageTimes, nil
}

// fetchFarmerCropsCount fetch farmer crops count
func (r *Repository) fetchFarmerCropsCount(farmIDs pq.StringArray, filters models.CropsFilters) (int, error) {
	var total int
	status := pq.Array(filters.CropStages)
	args := []interface{}{
		farmIDs,
		filters.SearchString,
		status,
		filters.SearchString == "",
	}

	SQL := `SELECT count(distinct fc.id)
			FROM farm_crops fc
					 JOIN crops c on c.id = fc.crop_id
					 LEFT JOIN crop_images ci on ci.farmer_crop_id = fc.id AND ci.crop_status = any($3)
			WHERE fc.farm_id = ANY($1)
			  AND fc.archived_at IS NULL
			  AND (true=$4 OR c.crop_name ILIKE '%%' || $2 || '%%')
			  AND fc.crop_stage = any($3)`

	err := r.Get(&total, SQL, args...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return total, errors.Wrap(err, "fetchFarmerCropsCount")
	}

	return total, nil
}

// addFarmCrop add farm crop to a farm
func (r *Repository) addFarmCrop(crop cropFormRequest) (string, error) {
	var cropID string
	args := []interface{}{
		crop.CropID,
		crop.FarmID,
		crop.Stage,
	}

	err := r.txx(func(txRepo *Repository) error {

		SQL := `INSERT INTO farm_crops (
                 crop_id,
    			 farm_id,
    			 crop_stage
    			 )
				VALUES ($1,$2,$3) RETURNING id`
		err := txRepo.Get(&cropID, SQL, args...)
		if err != nil {
			return errors.Wrap(err, "addFarmCrop")
		}

		SQL = `INSERT INTO farm_crop_stages (
                 stage,
    			 starting_time,
    			 farm_crop_id
    			 )
				VALUES ($1,$2,$3) RETURNING id`
		_, err = txRepo.Exec(SQL, crop.Stage, time.Now(), cropID)
		if err != nil {
			return errors.Wrap(err, "addFarmCrop")
		}

		if crop.Stage == models.Cropping || crop.Stage == models.Harvesting || crop.Stage == models.SunDrying {
			err = txRepo.editCropping(cropID, crop.CroppingDetails)
			if err != nil {
				return errors.Wrap(err, "addFarmCrop")
			}
		}

		if crop.Stage == models.SunDrying || crop.Stage == models.Harvesting {
			err = txRepo.editHarvesting(cropID, crop.HarvestingDetails)
			if err != nil {
				return errors.Wrap(err, "addFarmCrop")
			}
		}

		if crop.Stage == models.SunDrying {
			err = txRepo.editSundrying(cropID, crop.SundryingDetails)
			if err != nil {
				return errors.Wrap(err, "addFarmCrop")
			}
		}

		return nil
	})

	if err != nil {
		return cropID, errors.Wrap(err, "addFarmCrop")
	}
	return cropID, nil
}

// imagesAdd images add and archive previous
func (r *Repository) imagesAdd(stage models.StageCrop, farmCropId string, cropImageIDs []string) error {
	SQL := `UPDATE crop_images
			SET archived_at = now()
			WHERE farmer_crop_id = $1
			  and crop_status = $2
			AND archived_at IS NULL`
	_, err := r.Exec(SQL, farmCropId, stage)
	if err != nil {
		return errors.Wrap(err, "imagesAdd")
	}

	qBuilder := squirrel.Insert("crop_images").Columns("crop_status", "image_id", "farmer_crop_id")
	for idx := range cropImageIDs {
		qBuilder = qBuilder.Values(stage, cropImageIDs[idx], farmCropId)
	}
	finalSQL, args2, err := qBuilder.PlaceholderFormat(squirrel.Dollar).ToSql()
	_, err = r.Exec(finalSQL, args2...)
	if err != nil {
		return errors.Wrap(err, "imagesAdd")
	}
	return nil

}

// getFertilizers get admin fertilizers
func (r *Repository) getFertilizers(filters models.GenericFilters) ([]fertilizer, error) {
	args := []interface{}{
		filters.SearchString,
		filters.Limit,
		filters.Limit * filters.Page,
		filters.SearchString == "",
	}
	fertilizers := make([]fertilizer, 0)
	SQL := `SELECT id,
				name
			FROM fertilizers
			WHERE archived_at IS NULL
			 AND (true=$4 OR name ILIKE '%%' || $1 || '%%')
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`
	err := r.Select(&fertilizers, SQL, args...)
	if err != nil {
		return fertilizers, errors.Wrap(err, "getFertilizer")
	}
	return fertilizers, nil
}

func (r *Repository) getVideos(filters models.VideoFilters) ([]video, error) {
	videoType := "biochar"
	if filters.VideoType.Valid {
		videoType = filters.VideoType.String
	}
	args := []interface{}{
		filters.SearchString,
		filters.Limit,
		filters.Limit * filters.Page,
		filters.SearchString == "",
		!filters.VideoType.Valid,
		videoType,
	}
	videos := make([]video, 0)
	SQL := `SELECT 
    			video.id, 
    			video.title, 
    			video.description, 
    			video.url as video_url, 
    			video.video_tag, 
    			video.thumbnail_image_id,
    			uploads.path as thumbnail_image_path
			FROM video
			left join uploads on video.thumbnail_image_id = uploads.id
			WHERE video.archived_at IS NULL
			 AND (true=$4 OR video.title ILIKE '%%' || $1 || '%%')
			 AND (true=$5 OR video.video_tag = $6)
			ORDER BY video.created_at DESC
			LIMIT $2 OFFSET $3`
	err := r.Select(&videos, SQL, args...)
	if err != nil {
		return videos, errors.Wrap(err, "getVideos")
	}
	return videos, nil
}

// fetchCropsCount fetch crops count
func (r *Repository) fetchFertilizersCount(filters models.GenericFilters) (int, error) {
	var count int
	SQL := `SELECT COUNT(id)
			 FROM fertilizers
			 WHERE name ILIKE coalesce('%%' || $1 || '%%', name)
			  AND archived_at IS NULL`
	err := r.Get(&count, SQL, filters.SearchString)
	if err != nil {
		return 0, errors.Wrap(err, "fetchFertilizersCount")
	}
	return count, nil
}

// fetchCropsCount fetch crops count
func (r *Repository) fetchVideosCount(filters models.VideoFilters) (int, error) {
	var count int
	videoType := "biochar"
	if filters.VideoType.Valid {
		videoType = filters.VideoType.String
	}
	args := []interface{}{
		filters.SearchString,
		filters.SearchString == "",
		!filters.VideoType.Valid,
		videoType,
	}
	SQL := `SELECT COUNT(id)
			 FROM video
			 WHERE archived_at IS NULL
			     AND (true=$2 OR video.title ILIKE '%%' || $1 || '%%')
			     AND (true=$3 OR video.video_tag = $4)`
	err := r.Get(&count, SQL, args...)
	if err != nil {
		return 0, errors.Wrap(err, "fetchVideosCount")
	}
	return count, nil
}

// editCropping edit cropping
func (r *Repository) editCropping(farmCropId string, croppingRequest editCroppingRequest) error {
	args := []interface{}{
		croppingRequest.SeedQuantity,
		croppingRequest.Unit,
		farmCropId,
	}
	fertilizers := croppingRequest.Fertilizers

	SQL := `UPDATE farm_crops
					SET 
					    seed_quantity                = $1,
						seed_quantity_unit           = $2,
						updated_at          = now()
				WHERE id = $3`
	_, err := r.Exec(SQL, args...)
	if err != nil {
		return errors.Wrap(err, "editCropping")
	}

	err = r.imagesAdd(models.Cropping, farmCropId, croppingRequest.ImageIds) // new logic seperated in function

	SQL = `UPDATE crop_fertilizers
			SET archived_at = now()
			WHERE farm_crop_id = $1
			AND archived_at IS NULL`
	_, err = r.Exec(SQL, farmCropId)
	if err != nil {
		return errors.Wrap(err, "editCropping")
	}

	qBuilder := squirrel.Insert("crop_fertilizers").Columns("fertilizer_id", "fertilizer_quantity", "fertilizer_quantity_unit", "farm_crop_id")
	for idx := range fertilizers {
		qBuilder = qBuilder.Values(fertilizers[idx].Id, fertilizers[idx].Weight, fertilizers[idx].Unit, farmCropId)
	}
	finalSQL, args2, err := qBuilder.PlaceholderFormat(squirrel.Dollar).ToSql()
	_, err = r.Exec(finalSQL, args2...)
	if err != nil {
		return errors.Wrap(err, "editCropping")
	}
	return nil
}

// editHarvesting edit harvesting
func (r *Repository) editHarvesting(farmCropId string, harvestingRequest editHarvestingRequest) error {
	args := []interface{}{
		farmCropId,
		harvestingRequest.YieldQuantity,
		harvestingRequest.Unit,
	}

	SQL := `UPDATE farm_crops
				SET 
				    yield_quantity                = $2,
					yield_quantity_unit           = $3,
					updated_at          = now()
			    WHERE id = $1;`
	_, err := r.Exec(SQL, args...)
	if err != nil {
		return errors.Wrap(err, "ediHarvesting")
	}

	return r.imagesAdd(models.Harvesting, farmCropId, harvestingRequest.ImageIds) // new logic seperated in function
}

// editSundrying edit sun-drying
func (r *Repository) editSundrying(farmCropId string, crop editSundryingRequest) error {
	return r.imagesAdd(models.SunDrying, farmCropId, crop.ImageIds) // new logic seperated in function
}

// getFarmCropDetails
func (r *Repository) getFarmCropDetails(cropID string) (farmCropDetails, error) {
	var cropDetail farmCropDetails
	SQL := `SELECT farm_crops.id,
				   farms.id AS farm_id,
				   farms.landmark,
				   users.id AS farmer_id,
				   users.name AS farmer_name,
				   users.number,
				   users.country_code,
				   crops.crop_name,
				   crops.season,
				   crops.crop_image_id,
				   farm_crops.crop_area,
				   farm_crops.crop_area_unit,
				   farm_crops.crop_stage,
				   farm_crops.seed_quantity,
				   farm_crops.seed_quantity_unit,
				   farm_crops.yield_quantity,
				   farm_crops.yield_quantity_unit,
				   farm_crops.biomass_quantity,
				   farm_crops.biomass_quantity_unit,
				   farm_crops.biomass_transportation_vehicle_type,
				   klin.id AS kiln_id,
				   klin.name AS kiln_name,
				   farm_crops.biochar_quantity,
				   farm_crops.biochar_quantity_unit
			FROM users
					 JOIN farms on users.id = farms.farmer_id
					 JOIN farm_crops on farm_crops.farm_id = farms.id
					 JOIN crops on crops.id = farm_crops.crop_id
					 LEFT JOIN klin on farm_crops.klin_id = klin.id
			WHERE farm_crops.id = $1
			  AND farm_crops.archived_at IS NULL 
			GROUP BY farm_crops.id, crops.id, farms.id, users.id, klin.id`
	err := r.Get(&cropDetail, SQL, cropID)
	if err != nil {
		return cropDetail, errors.Wrap(err, "GetFarmCropDetails")
	}
	return cropDetail, nil
}

// fetchFarmCropStages fetch Farmer crops stages
func (r *Repository) fetchFarmCropStages(farmCropId string) ([]cropStageTime, error) {
	cropStageTimes := make([]cropStageTime, 0)

	SQL := `SELECT fcs.id,
				   fcs.stage,
				   fcs.starting_time
			FROM farm_crop_stages fcs
			WHERE fcs.farm_crop_id = $1`
	err := r.Select(&cropStageTimes, SQL, farmCropId)
	if err != nil {
		return nil, errors.Wrap(err, "fetchFarmerCropStages")
	}

	return cropStageTimes, nil
}

// fetchFarmCropFertilizers fetch Farmer crops fertilizers
func (r *Repository) fetchFarmCropFertilizers(farmCropID string) ([]fertilizerInfo, error) {
	fertilizersInfo := make([]fertilizerInfo, 0)

	SQL := `SELECT DISTINCT ON (fertilizers.name) crop_fertilizers.fertilizer_id,
       			   fertilizers.name,
				   crop_fertilizers.fertilizer_quantity,
				   crop_fertilizers.fertilizer_quantity_unit
			FROM crop_fertilizers 
				 Join fertilizers on crop_fertilizers.fertilizer_id = fertilizers.id
			WHERE crop_fertilizers.farm_crop_id = $1`

	err := r.Select(&fertilizersInfo, SQL, farmCropID)
	if err != nil {
		return nil, errors.Wrap(err, "fetchFarmerCropFertilizers")
	}

	return fertilizersInfo, nil
}

// cropChangeStatus crop Change Status
func (r *Repository) cropChangeStatus(farmCropId string, cropStatus models.StageCrop) error {

	SQL := `UPDATE farm_crops
				SET crop_stage =$1
			WHERE id = $2
			AND archived_at IS NULL`

	_, err := r.Exec(SQL, cropStatus, farmCropId)
	if err != nil {
		return errors.Wrap(err, "cropChangeStatus")
	}

	SQL = `INSERT INTO farm_crop_stages (
                 stage,
    			 starting_time,
    			 farm_crop_id
    			 )
				VALUES ($1,$2,$3) RETURNING id`
	_, err = r.Exec(SQL, cropStatus, time.Now(), farmCropId)
	if err != nil {
		return errors.Wrap(err, "cropChangeStatus")
	}

	return nil
}

// getImageStates get image states
func (r *Repository) getImageStates(farmCropID string) (cropStages, error) {
	stages := make([]models.StageCrop, 0)
	cropStages := cropStages{
		Stages: stages,
	}
	SQL := `SELECT crop_status
			FROM crop_images
			WHERE farmer_crop_id = $1
			AND archived_at IS NULL 
			GROUP BY crop_status`

	err := r.Select(&stages, SQL, farmCropID)
	cropStages.Stages = stages
	if err != nil {
		return cropStages, errors.Wrap(err, "getImagesStates")
	}

	return cropStages, nil
}

// getCropImageDetails fetch image path
func (r *Repository) getCropImageDetails(farmCropID string) (images []imageDetails, err error) {
	SQL := `SELECT DISTINCT ON (uploads.id) uploads.id,
       				uploads.path,
				   crop_images.crop_status
			FROM crop_images
			JOIN uploads ON crop_images.image_id = uploads.id
			WHERE crop_images.farmer_crop_id = $1`
	err = r.Select(&images, SQL, farmCropID)
	if err != nil {
		return images, err
	}
	return images, nil
}

// fetchUserPreferredCrops fetch user preferred crops
func (r *Repository) fetchUserPreferredCrops(farmerID string) (cropsInfo []preferredCrops, err error) {
	SQL := `SELECT crops.id,
				   crops.crop_name,
				   crops.crop_image_id
			FROM crops
			JOIN farmer_preferred_crop on crops.id = farmer_preferred_crop.crop_id
				WHERE farmer_preferred_crop.farmer_id = $1
				  AND farmer_preferred_crop.archived_at IS NULL`
	err = r.Select(&cropsInfo, SQL, farmerID)
	if err != nil {
		return cropsInfo, err
	}
	return cropsInfo, nil
}

// fetchSingleImagePath fetch single image path
func (r *Repository) fetchSingleImagePath(imageID string) (images image, err error) {
	SQL := `SELECT id,
				   path,
				   type
			FROM uploads
			WHERE id = $1
			LIMIT 1`
	err = r.Get(&images, SQL, imageID)
	if err != nil {
		return images, err
	}
	return images, nil
}

// editTransportation edit transportation
func (r *Repository) editTransportation(farmCropId string, transportationRequest editTransportationRequest) error {
	args := []interface{}{
		farmCropId,
		string(transportationRequest.VehicleType),
	}

	SQL := `UPDATE farm_crops
				SET 
				    biomass_transportation_vehicle_type = $2,
					updated_at = now()
			WHERE id = $1;`
	_, err := r.Exec(SQL, args...)
	if err != nil {
		return errors.Wrap(err, "editTransportation")
	}

	return r.imagesAdd(models.Transportation, farmCropId, transportationRequest.ImageIds) // new logic seperated in function
}

// addMoveToProductionCropDetails add Move To Production Crop Details
func (r *Repository) addMoveToProductionCropDetails(farmCropId string, crop moveToProductionRequest) error {
	args := []interface{}{
		farmCropId,
		crop.KilnID,
		crop.BiomassQuantity,
		crop.BiomassQuantityUnit,
	}

	SQL := `UPDATE farm_crops
				SET 
				    klin_id = $2,
				    biomass_quantity = $3,
				    biomass_quantity_unit =$4,
					updated_at = now()
				WHERE id = $1;`
	_, err := r.Exec(SQL, args...)
	if err != nil {
		return errors.Wrap(err, "addKilnToFarmCrop")
	}

	return nil
}

// editDistribution edit distribution
func (r *Repository) editDistribution(farmCropId string, crop editDistributionRequest) error {
	return r.imagesAdd(models.Distribution, farmCropId, crop.ImageIds) // new logic seperated in function
}
