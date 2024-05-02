package kilnOperator

import (
	"circonomy-server/dbutil"
	"circonomy-server/models"
	"circonomy-server/repobase"
	"database/sql"
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

// txx transaction
func (r *Repository) txx(fn func(txRepo *Repository) error) error {
	return dbutil.WithTransaction(r.DB(), func(tx *sqlx.Tx) error {
		repoCopy := *r
		repoCopy.Base = r.Base.CopyWithTX(tx)
		return fn(&repoCopy)
	})
}

// updateOTP delete previous OTP and send otp to user
func (r *Repository) updateOTP(otpReq models.SendOTP) error {
	return r.txx(func(txRepo *Repository) error {
		err := txRepo.deletePreviousOTP(otpReq)
		if err != nil {
			return err
		}
		return txRepo.sendOTP(otpReq)
	})
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

// checkKilnOperator check kiln operator
func (r *Repository) checkKilnOperator(otpReq sendOTPRequest) error {
	var count int
	SQL := `SELECT count(u.id) FROM 
             	 klin_operator 
         	JOIN users u on u.id = klin_operator.operator_id
			WHERE u.number = $1
			  AND u.country_code = $2
			  AND u.archived_at is null
			  AND klin_operator.archived_at IS NULL`
	err := r.Get(&count, SQL, otpReq.PhoneNumber, otpReq.CountryCode)
	if err != nil {
		return errors.Wrap(err, "checkKilnOperator")
	}
	if count == 0 {
		return errors.New("kiln operator not present")
	}
	return nil
}

// sendOTP send OTP
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

// fetchUserID
func (r *Repository) fetchUserID(phoneNumber string, countryCode string) (string, error) {
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

// verifyOTP verify correct otp present in db
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

// getFarmCropsBiomasses get farm crops biomasses
func (r *Repository) getFarmCropsBiomasses(kilnOperatorID string, filters models.KilnFilters) ([]farmCropBiomasses, error) {
	cropsFiltered := make([]farmCropBiomasses, 0)

	args := []interface{}{
		kilnOperatorID,
		filters.Limit,
		filters.Limit * filters.Page,
		string(models.TransportFarmToKiln),
		filters.KilnID.String,
	}

	SQL := `SELECT farm_crops.id,
       			   users.name AS farmer_name,
       			   users.number,
       			   users.country_code,
				   crops.crop_name,
				   farm_crops.biomass_quantity,
				   farm_crops.biomass_quantity_unit,
				   array_remove(ARRAY_AGG(distinct uploads.id), NULL) AS farm_crop_images_ids
			FROM farm_crops
					 JOIN crops on crops.id = farm_crops.crop_id
			    	 JOIN klin on farm_crops.klin_id = klin.id
			    	 JOIN klin_operator on klin.id = klin_operator.klin_id
			    	 JOIN farms on farm_crops.farm_id = farms.id
			    	 JOIN users on farms.farmer_id = users.id
					 LEFT JOIN crop_images on crop_images.farmer_crop_id = farm_crops.id AND crop_images.crop_status = $4
					 LEFT JOIN uploads on uploads.id = crop_images.image_id
			WHERE klin_operator.operator_id = $1
			  AND farm_crops.archived_at IS NULL
			  AND farm_crops.crop_stage = $4
			  AND klin.id = $5
			  AND crop_images.archived_at IS NULL 
			  AND farm_crops.biomass_verified_kiln_operator = false
			GROUP BY farm_crops.id, crops.crop_name, farm_crops.updated_at, users.id
			ORDER BY farm_crops.updated_at DESC
			LIMIT $2 OFFSET $3`

	err := r.Select(&cropsFiltered, SQL, args...)
	if err != nil {
		return cropsFiltered, errors.Wrap(err, "getFarmCropsBiomasses")
	}
	return cropsFiltered, nil
}

// getFarmCropsBiomassesCount fetch farmer crops count
func (r *Repository) getFarmCropsBiomassesCount(kilnOperatorID string) (int, error) {
	var total int
	status := string(models.TransportFarmToKiln)
	args := []interface{}{
		kilnOperatorID,
		status,
	}

	SQL := `SELECT count(farm_crops.id)
			FROM farm_crops
					 JOIN crops on crops.id = farm_crops.crop_id
			    	 JOIN klin on farm_crops.klin_id = klin.id
			    	 JOIN klin_operator on klin.id = klin_operator.klin_id
			    	 JOIN farms on farm_crops.farm_id = farms.id
			    	 JOIN users on farms.farmer_id = users.id
			WHERE klin_operator.operator_id = $1
			  AND farm_crops.archived_at IS NULL
			  AND farm_crops.crop_stage = $2
			  AND farm_crops.biomass_verified_kiln_operator = false`

	err := r.Get(&total, SQL, args...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return total, errors.Wrap(err, "fetchFarmerCropsCount")
	}

	return total, nil
}

// fetchImagePath fetch image path
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

// fetchImagePath fetch image path
func (r *Repository) fetchImagePath(imageIDs pq.StringArray) (images []image, err error) {
	SQL := `SELECT id,
				   path,
				   type
			FROM uploads
			WHERE id = ANY ($1)`
	err = r.Select(&images, SQL, imageIDs)
	if err != nil {
		return images, err
	}
	return images, nil
}

// fetchImagePath fetch image path
func (r *Repository) fetchKlinVideoPath(klinProcessIDs []string) (videos []video, err error) {
	SQL := `SELECT uploads.id,
       			   klin_process_images.klin_process_id,
				   path
			FROM uploads
			JOIN klin_process_images on uploads.id = klin_process_images.upload_id
			WHERE klin_process_images.id = any($1)
			AND klin_process_images.file_type = $2`
	err = r.Select(&videos, SQL, pq.StringArray(klinProcessIDs), videoFileType)
	if err != nil {
		return videos, err
	}
	return videos, nil
}

// editFarmCropAndAddKilnBiomass
func (r *Repository) editFarmCropAndAddKilnBiomass(farmCropId, kilnID string, croppingRequest editFarmCropBiomassRequest) error {
	args := []interface{}{
		croppingRequest.BiomassQuantity,
		croppingRequest.Unit,
		farmCropId,
		models.Production,
		kilnID,
	}
	SQL := `UPDATE farm_crops
				SET 
				    biomass_quantity                = $1,
					biomass_quantity_unit           = $2,
					biomass_verified_kiln_operator = true,
					updated_at          		   = now(),
					crop_stage 					   = $4,
					klin_id 					   = $5
			WHERE id = $3
			RETURNING crop_id`
	var cropID string
	err := r.Get(&cropID, SQL, args...)
	if err != nil {
		return errors.Wrap(err, "editFarmCropAndAddKilnBiomass")
	}

	err = r.imagesAdd(models.TransportFarmToKiln, farmCropId, croppingRequest.ImageIds)
	if err != nil {
		return errors.Wrap(err, "editFarmCropAndAddKilnBiomass")
	}

	err = r.checkKilnBiomassAndInsertRow(cropID, kilnID)
	if err != nil {
		return errors.Wrap(err, "editFarmCropAndAddKilnBiomass")
	}

	args = []interface{}{
		kilnID,
		cropID,
		croppingRequest.BiomassQuantity,
	}
	SQL = `UPDATE klin_biomass SET
    	   		current_quantity = current_quantity + $3
           WHERE klin_id = $1
           AND crop_id = $2`
	_, err = r.Exec(SQL, args...)
	if err != nil {
		return errors.Wrap(err, "editFarmCropAndAddKilnBiomass")
	}

	return nil
}

// checkKilnBiomassAndInsertRow
func (r *Repository) checkKilnBiomassAndInsertRow(cropID, kilnID string) error {

	kilnBiomass := kilnBiomass{}
	SQL := `SELECT klin_biomass.klin_id AS kiln_id,
				  crops.Id AS crop_id,
				  crops.crop_name,
				  klin_biomass.current_quantity
			FROM klin_biomass
			join crops on crops.id = klin_biomass.crop_id
			WHERE klin_id = $1
			and crop_id = $2`
	err := r.Get(&kilnBiomass, SQL, kilnID, cropID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return errors.Wrap(err, "editFarmCropAndAddKilnBiomass")
	}
	if errors.Is(err, sql.ErrNoRows) {
		args := []interface{}{
			kilnID,
			cropID,
			0,
			time.Now(),
		}
		SQL = `INSERT INTO klin_biomass
    	   		       (klin_id, crop_id, current_quantity, created_at)
				VALUES ($1, $2, $3, $4)`
		_, err = r.Exec(SQL, args...)
		if err != nil {
			return errors.Wrap(err, "editFarmCropAndAddKilnBiomass")
		}
	}
	return nil
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

// fetchKilnOperatorInfo fetch kiln operator info
func (r *Repository) fetchKilnOperatorInfo(kilnOperatorID string) (farmerInfo kilnOperatorDetails, err error) {

	SQL := `SELECT users.id,
				   users.name,
				   users.age,
				   users.gender,
				   users.address,
				   users.number,
				   users.country_code,
				   users.upload_id AS profile_image_id,
				   users.aadhaar_no,
				   users.aadhaar_no_image_id
			FROM users
			WHERE users.id = $1
			  	AND users.archived_at IS NULL
			GROUP BY users.id`

	err = r.Get(&farmerInfo, SQL, kilnOperatorID)
	if err != nil {
		return farmerInfo, err
	}
	return farmerInfo, nil
}

// fetchKilnInfo fetch kiln info
func (r *Repository) fetchKilnInfo(kilnOperatorID string) (kilns []kilnInfo, err error) {
	SQL := `SELECT klin.id,
				   klin.name,
				   klin.network_id,
				   network.name AS network_name,
				   klin.address,
				   klin.biochar_quantity
			FROM klin
				JOIN klin_operator on klin.id = klin_operator.klin_id
				JOIN users on klin_operator.operator_id = users.id
				JOIN network on klin.network_id = network.id
			WHERE users.id = $1
			AND users.archived_at is null
			AND klin_operator.archived_at is null
			AND klin.archived_at is null`
	err = r.Select(&kilns, SQL, kilnOperatorID)
	if err != nil {
		return kilns, err
	}
	return kilns, nil
}

// getKilnCropBiomasses get Kiln Crop Biomasses
func (r *Repository) getKilnCropBiomasses(kilnID string) ([]kilnBiomass, error) {
	kilnBiomasses := make([]kilnBiomass, 0)

	SQL := `SELECT klin_biomass.klin_id AS kiln_id,
				  crops.Id AS crop_id,
				  crops.crop_name,
				  klin_biomass.current_quantity
			FROM klin_biomass
			left join crops on crops.id = klin_biomass.crop_id
			WHERE klin_id = $1`
	err := r.Select(&kilnBiomasses, SQL, kilnID)
	if err != nil {
		return kilnBiomasses, errors.Wrap(err, "getKilnCropBiomasses")
	}
	return kilnBiomasses, nil
}

// getKilnProcessDetails get Kiln Ongoing Process Details
func (r *Repository) getKilnProcessDetails(kilnID string, filters models.GenericFilters) (kilnProcessResponse, error) {
	kilnProcesses := kilnProcessResponse{}

	SQL := `SELECT DISTINCT ON (klin_process.end_time) klin_process.id,
       			   klin_process.klin_id AS kiln_id,
				   crops.Id AS crop_id,
				   crops.crop_name,
				   klin_process.biochar_quantity,
				   klin_process.biomass_quantity,
				   klin_process.created_at,
				   klin_process.created_at AS starting_date,
				   klin_process.end_time,
				   array_remove(ARRAY_AGG(distinct uploads.id), NULL) AS kiln_process_images_ids
			FROM klin_process
			 join crops on crops.id = klin_process.crop_id
			 left join klin_process_images kpi on klin_process.id = kpi.klin_process_id and kpi.archived_at is null
			 LEFT JOIN uploads on uploads.id = kpi.upload_id
			WHERE klin_id = $1
			and klin_process.archived_at is null
			GROUP BY klin_process.id, crops.id, klin_process.end_time
				ORDER BY klin_process.end_time DESC
			LIMIT $2 OFFSET $3`

	err := r.Select(&kilnProcesses.KilnProcesses, SQL, kilnID, filters.Limit, filters.Page*filters.Limit)
	if err != nil {
		return kilnProcesses, errors.Wrap(err, "getKilnCropBiomasses")
	}
	kilnProcesses.Limit = filters.Limit
	kilnProcesses.Page = filters.Page
	return kilnProcesses, nil
}

// getKilnProcessDetails get Kiln Ongoing Process Details
func (r *Repository) getKilnProcessDetailsCount(kilnID string) (int, error) {
	SQL := `SELECT count(DISTINCT(klin_process.id))
			FROM klin_process
			WHERE klin_id = $1
			and klin_process.archived_at is null
			`

	var count int
	err := r.Get(&count, SQL, kilnID)
	return count, err
}

// addKilnProcessBatch add kiln process batch
func (r *Repository) addKilnProcessBatch(kilnId string, request kilnProcessRequest) error {
	var kilnProcessID string

	SQL := `INSERT INTO klin_process (klin_id, biomass_quantity, crop_id)
			VALUES ($1,$2,$3) returning id`
	err := r.Get(&kilnProcessID, SQL, kilnId, request.BiomassQuantity, request.CropID)
	if err != nil {
		return err
	}

	SQL = `UPDATE klin_biomass
			SET current_quantity = current_quantity - $1
			WHERE crop_id = $2`
	_, err = r.Exec(SQL, request.BiomassQuantity, request.CropID)
	if err != nil {
		return errors.Wrap(err, "addKilnProcessBatch, updating kiln biomass quantity")
	}

	err = r.kilnImagesAdd(kilnProcessID, request.ImageIds)
	if err != nil {
		return errors.Wrap(err, "addKilnProcessBatch")
	}

	err = r.kilnVideosAdd(kilnProcessID, request.VideoIds)
	if err != nil {
		return errors.Wrap(err, "addKilnProcessBatch")
	}

	return nil
}

// imagesAdd images add and archive previous
func (r *Repository) kilnImagesAdd(kilnProcessId string, kilnImageIds []string) error {
	SQL := `UPDATE klin_process_images
			SET archived_at = now()
			WHERE klin_process_id = $1
			AND file_type = 'image'
			AND archived_at IS NULL`
	_, err := r.Exec(SQL, kilnProcessId)
	if err != nil {
		return errors.Wrap(err, "kilnImagesAdd")
	}

	qBuilder := squirrel.Insert("klin_process_images").Columns("upload_id", "klin_process_id", "file_type")
	for idx := range kilnImageIds {
		qBuilder = qBuilder.Values(kilnImageIds[idx], kilnProcessId, imageFileType)
	}
	finalSQL, args2, err := qBuilder.PlaceholderFormat(squirrel.Dollar).ToSql()
	_, err = r.Exec(finalSQL, args2...)
	if err != nil {
		return errors.Wrap(err, "kilnImagesAdd")
	}
	return nil
}

// imagesAdd images add and archive previous
func (r *Repository) kilnVideosAdd(kilnProcessId string, kilnImageIds []string) error {
	SQL := `UPDATE klin_process_images
			SET archived_at = now()
			WHERE klin_process_id = $1
			AND file_type = 'video'
			AND archived_at IS NULL`
	_, err := r.Exec(SQL, kilnProcessId)
	if err != nil {
		return errors.Wrap(err, "kilnImagesAdd")
	}

	qBuilder := squirrel.Insert("klin_process_images").Columns("upload_id", "klin_process_id", "file_type")
	for idx := range kilnImageIds {
		qBuilder = qBuilder.Values(kilnImageIds[idx], kilnProcessId, videoFileType)
	}
	finalSQL, args2, err := qBuilder.PlaceholderFormat(squirrel.Dollar).ToSql()
	_, err = r.Exec(finalSQL, args2...)
	if err != nil {
		return errors.Wrap(err, "kilnImagesAdd")
	}
	return nil
}

// editKilnProcessDetails edit kiln process details
func (r *Repository) editKilnProcessDetails(kilnId string, request kilnProcessEditRequest) error {
	var count int
	SQL := `SELECT count(*)
			FROM klin_process
			WHERE klin_id = $1
			AND id = $2`
	err := r.Get(&count, SQL, kilnId, request.KilnProcessID)
	if err != nil {
		return errors.Wrap(err, "getKilnCropBiomasses")
	}
	if count != 0 {
		err := r.kilnImagesAdd(request.KilnProcessID, request.ImageIds)
		if err != nil {
			return errors.Wrap(err, "editKilnProcessDetails")
		}

		err = r.kilnVideosAdd(request.KilnProcessID, request.VideoIds)
		if err != nil {
			return errors.Wrap(err, "addKilnProcessBatch")
		}

		return nil
	}
	return errors.New("kiln process not under the relevant kiln Id")
}

// doneKilnProcess done kiln process
func (r *Repository) doneKilnProcess(kilnId string, request kilnProcessDoneRequest) error {
	args := []interface{}{
		time.Now(),
		request.KilnProcessID,
		request.BioCharQuantity,
	}
	SQL := `UPDATE klin_process
			SET end_time = $1,
			    biochar_quantity =$3
			WHERE id = $2`
	_, err := r.Exec(SQL, args...)
	if err != nil {
		return errors.Wrap(err, "doneKilnProcess")
	}

	SQL = `UPDATE klin
			SET biochar_quantity = coalesce(biochar_quantity,0) + $1
			WHERE id = $2`
	_, err = r.Exec(SQL, request.BioCharQuantity, kilnId)
	if err != nil {
		return errors.Wrap(err, "doneKilnProcess")
	}

	err = r.kilnImagesAdd(request.KilnProcessID, request.ImageIds)
	if err != nil {
		return errors.Wrap(err, "doneKilnProcess")
	}

	err = r.kilnVideosAdd(request.KilnProcessID, request.VideoIds)
	if err != nil {
		return errors.Wrap(err, "addKilnProcessBatch")
	}
	return nil
}

// fetchKilnBiochar fetch kiln biochar
func (r *Repository) fetchKilnBiochar(kilnId string) (kilnBioChar, error) {
	var biochar kilnBioChar
	SQL := `SELECT biochar_quantity 
			FROM klin
			WHERE id = $1
			and archived_at is null`
	err := r.Get(&biochar, SQL, kilnId)
	if err != nil {
		return biochar, errors.Wrap(err, "doneKilnProcess")
	}
	return biochar, nil
}

// addFarmCropBioCharDetails add farm crop biochar details
func (r *Repository) addFarmCropBioCharDetails(farmCropId string, request kilnDistributionRequest) error {
	var kilnProcessID string
	args := []interface{}{
		request.BioCharQuantity,
		farmCropId,
		ton,
		models.TransportKilnToFarm,
	}
	SQL := `UPDATE farm_crops
			SET biochar_quantity =$1,
				biochar_quantity_unit = $3,
				crop_stage =$4
			WHERE id = $2`
	_, err := r.Exec(SQL, args...)
	if err != nil {
		return errors.Wrap(err, "addFarmCropBioCharDetails")
	}

	SQL = `UPDATE klin
			SET biochar_quantity = biochar_quantity - $1
			WHERE id = $2`
	_, err = r.Exec(SQL, request.BioCharQuantity, request.KilnID)
	if err != nil {
		return errors.Wrap(err, "addFarmCropBioCharDetails")
	}

	SQL = `SELECT klin_process.id 
			FROM klin_process
			join farm_crops fc on klin_process.klin_id = fc.klin_id
		    WHERE klin_process.klin_id = $1
		    AND fc.id =$2`
	err = r.Get(&kilnProcessID, SQL, request.KilnID, farmCropId)
	if err != nil {
		return errors.Wrap(err, "addFarmCropBioCharDetails")
	}

	SQL = `INSERT INTO farm_crop_stages (stage, starting_time, farm_crop_id)
			VALUES ($1,$2,$3)`
	_, err = r.Exec(SQL, models.TransportKilnToFarm, time.Now(), farmCropId)
	if err != nil {
		return err
	}

	err = r.imagesAdd(models.TransportKilnToFarm, farmCropId, request.ImageIds)
	if err != nil {
		return errors.Wrap(err, "addFarmCropBioCharDetails")
	}
	return nil
}

// fetchDistributedFarmCrops fetch distributed farm crops
func (r *Repository) fetchDistributedFarmCrops(kilnID string, filters models.GenericFilters) ([]distributedFarmCrop, error) {
	farmCrop := make([]distributedFarmCrop, 0)
	args := []interface{}{
		kilnID,
		filters.Limit,
		filters.Limit * filters.Page,
		models.Distribution,
		models.TransportKilnToFarm,
	}
	SQL := `SELECT DISTINCT ON (farm_crops.id) 
    			   farm_crops.id,
				   users.name AS farmer_name,
				   users.number,
				   users.country_code,
				   farm_crops.biomass_quantity,
				   farm_crops.biomass_quantity_unit,
				   farm_crops.biochar_quantity,
				   farm_crops.biochar_quantity_unit,
				   farm_crop_stages.starting_time,
				   farm_crops.updated_at,
				   array_remove(ARRAY_AGG(distinct uploads.id), NULL) AS farm_images_ids
			FROM users
					 JOIN farms on users.id = farms.farmer_id
					 JOIN farm_crops on farm_crops.farm_id = farms.id
			    	 JOIN farm_crop_stages on farm_crops.id = farm_crop_stages.farm_crop_id
					 LEFT JOIN klin on farm_crops.klin_id = klin.id
					 LEFT JOIN crop_images on crop_images.farmer_crop_id = farm_crops.id AND crop_images.crop_status = $4
					 LEFT JOIN uploads on uploads.id = crop_images.image_id
			WHERE klin.id = $1
			  AND farm_crop_stages.stage = $4 OR farm_crop_stages.stage = $5
			  AND farm_crops.archived_at IS NULL 
			GROUP BY farm_crops.id, farms.id, users.id, klin.id, farm_crops.updated_at, farm_crop_stages.starting_time
			ORDER BY farm_crop_stages.starting_time DESC
			LIMIT $2 OFFSET $3`
	err := r.Select(&farmCrop, SQL, args...)
	if err != nil {
		return farmCrop, errors.Wrap(err, "GetFarmCropDetails")
	}
	return farmCrop, nil
}

// fetchDistributedFarmCrops fetch distributed farm crops
func (r *Repository) fetchDistributedFarmCropsCount(kilnID string) (int, error) {
	args := []interface{}{
		kilnID,
		models.TransportKilnToFarm,
	}
	SQL := `SELECT count(distinct(farm_crops.id))
			FROM farm_crops
			    	 JOIN farm_crop_stages on farm_crops.id = farm_crop_stages.farm_crop_id
			WHERE farm_crops.klin_id = $1
			  AND farm_crop_stages.stage = $2
			  AND farm_crops.archived_at IS NULL
	`
	var count int
	err := r.Get(&count, SQL, args...)
	if err != nil {
		return count, errors.Wrap(err, "fetchDistributedFarmCropsCount")
	}
	return count, nil
}
