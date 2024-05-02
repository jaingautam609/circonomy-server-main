package admin

import (
	"circonomy-server/dbutil"
	"circonomy-server/models"
	"circonomy-server/repobase"
	"circonomy-server/utils"
	"fmt"
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

// createCrop create crop
func (r *Repository) createCrop(crop AddCropRequest) (string, error) {
	var cropID string
	args := []interface{}{
		crop.Name,
		crop.Season,
		crop.ImageID,
	}

	SQL := `INSERT INTO crops (
                 crop_name,
    			 season,
    			 crop_image_id
    			 )
				VALUES ($1,$2,$3) RETURNING id`
	err := r.Get(&cropID, SQL, args...)
	if err != nil {
		return cropID, errors.Wrap(err, "createCrop")
	}
	return cropID, nil
}

// deleteCrop delete crop by id
func (r *Repository) deleteCrop(cropID string) error {
	var id string
	SQL := `SELECT id FROM crops
			WHERE id = $1
			AND archived_at IS NULL`
	err := r.Get(&id, SQL, cropID)
	if err != nil {
		return errors.Wrap(err, "delete crop, can't find crop with the id")
	}

	SQL = `UPDATE crops
			SET archived_at = now()
			WHERE id = $1
			AND archived_at IS NULL`
	_, err = r.Exec(SQL, cropID)
	if err != nil {
		return errors.Wrap(err, "deleteCrop")
	}
	return nil
}

// createCrop create crop
func (r *Repository) createVideo(video AddVideoRequest) (string, error) {
	var cropID string
	args := []interface{}{
		video.Title,
		video.Description,
		video.VideoURL,
		video.VideoType,
		video.ImageID,
	}

	SQL := `INSERT INTO video (
                title, 
                description, 
                url, 
                video_tag, 
                thumbnail_image_id
   			 )
				VALUES ($1,$2,$3, $4, $5) RETURNING id`
	err := r.Get(&cropID, SQL, args...)
	if err != nil {
		return cropID, errors.Wrap(err, "createVideo")
	}
	return cropID, nil
}

// deleteCrop delete crop by id
func (r *Repository) deleteVideo(videoID string) error {
	SQL := `UPDATE video
			SET archived_at = now()
			WHERE id = $1
			AND archived_at IS NULL`
	_, err := r.Exec(SQL, videoID)
	if err != nil {
		return errors.Wrap(err, "deleteVideo")
	}
	return nil
}

// createFertilizer create Fertilizer
func (r *Repository) createFertilizer(fertilizer AddFertilizerRequest) (string, error) {
	var fertilizerID string
	args := []interface{}{
		fertilizer.Name,
	}

	SQL := `INSERT INTO fertilizers (
            name
    		)
			VALUES ($1) RETURNING id`
	err := r.Get(&fertilizerID, SQL, args...)
	if err != nil {
		return fertilizerID, errors.Wrap(err, "createFertilizer")
	}
	return fertilizerID, nil
}

// deleteFertilizer delete fertilizer by id
func (r *Repository) deleteFertilizer(fertilizerID string) error {
	var id string
	SQL := `SELECT id FROM fertilizers
			WHERE id = $1
			AND archived_at IS NULL`
	err := r.Get(&id, SQL, fertilizerID)
	if err != nil {
		return errors.Wrap(err, "delete fertilizer, can't find fertilizer with the id")
	}

	SQL = `UPDATE fertilizers
			SET archived_at = now()
			WHERE id = $1
			AND archived_at IS NULL`
	_, err = r.Exec(SQL, fertilizerID)
	if err != nil {
		return errors.Wrap(err, "deleteFertilizer")
	}
	return nil
}

func (r *Repository) getUserPasswordByEmailAndAccountType(email string, accountType models.UserAccountType) (string, string, error) {
	// language = SQL
	SQL := `SELECT id, password from users where email = $1 AND archived_at IS NULL and account_type = $2`
	var id, pwd string
	err := r.QueryRowX(SQL, email, accountType).Scan(&id, &pwd)
	return id, pwd, err
}

// createUser create User
func (r *Repository) createUser(user user, accountType models.UserAccountType) (string, error) {

	var userID string
	var password string

	password, err := utils.HashPassword(user.Password.String)
	if err != nil {
		return userID, errors.Wrap(err, "createBiomassAggregator")
	}

	args := []interface{}{
		user.Name,
		user.Email,
		password,
		user.PhoneNo,
		user.CountryCode,
		accountType,
	}

	SQL := `INSERT INTO users (
                 name,
    			 email,
                 password,
                 number,
                 country_code,
                 account_type
    		)
			VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`
	err = r.Get(&userID, SQL, args...)
	if err != nil {
		return userID, errors.Wrap(err, "createUser-adminPackage")
	}

	return userID, nil
}

// createBiomassAggregator create biomass aggregator
func (r *Repository) createBiomassAggregator(biomassAggregator biomassAggregatorRequest) (string, error) {

	var biomassAggregatorID string
	location := fmt.Sprintf("(%v,%v)", biomassAggregator.Latitude, biomassAggregator.Longitude)
	args := []interface{}{
		biomassAggregator.Name,
		location,
		biomassAggregator.LocationName,
		biomassAggregator.Manager.Trained,
	}
	SQL := `INSERT INTO biomass_aggregator (
                 name,
    			 location,
                 location_name,
                 trained
    	   )
		   VALUES ($1,$2,$3,$4) RETURNING id`
	err := r.Get(&biomassAggregatorID, SQL, args...)
	if err != nil {
		return biomassAggregatorID, errors.Wrap(err, "createBiomassAggregator")
	}

	return biomassAggregatorID, nil
}

// createBiomassAggregatorManager create biomass aggregator managaer
func (r *Repository) createBiomassAggregatorManager(biomassAggregatorID string, managerID string) error {

	SQL := `INSERT INTO biomass_aggregator_manager (
                 biomass_aggregator_id,
    			 manager_id
    			 )
				VALUES ($1,$2)`
	_, err := r.Exec(SQL, biomassAggregatorID, managerID)
	if err != nil {
		return errors.Wrap(err, "createBiomassAggregator")
	}
	return nil
}

// getBiomassAggregator get biomass aggregator
func (r *Repository) getBiomassAggregator(filters models.GenericFilters) ([]biomassAggregator, error) {
	args := []interface{}{
		filters.Limit,
		filters.Limit * filters.Page,
		filters.SearchString,
		filters.SearchString == "",
	}
	biomassAggregators := make([]biomassAggregator, 0)
	SQL := `SELECT biomass_aggregator.id,
				   biomass_aggregator.name,
				   u.email,
				   count(distinct network.id) AS c_sink_count,
				   count(distinct farmer_details.user_id) AS farmers_count,
				   biomass_aggregator.trained,
				   biomass_aggregator.location_name
			FROM biomass_aggregator
				 left join network on biomass_aggregator.id = network.biomass_aggregator_id
				 left join farmer_details on network.id = farmer_details.network_id
			     left join biomass_aggregator_manager bam on biomass_aggregator.id = bam.biomass_aggregator_id
			     left join users u on bam.manager_id = u.id
			WHERE biomass_aggregator.archived_at IS NULL
				AND (true=$4 OR biomass_aggregator.name ILIKE '%%' || $3 || '%%')
			GROUP BY biomass_aggregator.id, biomass_aggregator.created_at, u.id
			ORDER BY biomass_aggregator.created_at DESC
			LIMIT $1 OFFSET $2`
	err := r.Select(&biomassAggregators, SQL, args...)
	if err != nil {
		return biomassAggregators, errors.Wrap(err, "getBiomassAggregator")
	}
	return biomassAggregators, nil
}

// getBiomassAggregatorCount get biomass aggregator count
func (r *Repository) getBiomassAggregatorCount(filters models.GenericFilters) (int, error) {
	var count int
	args := []interface{}{
		filters.SearchString,
		filters.SearchString == "",
	}
	SQL := `SELECT COUNT(distinct biomass_aggregator.id)
			 FROM biomass_aggregator
				 left join network on biomass_aggregator.id = network.biomass_aggregator_id
				 left join farmer_details on network.id = farmer_details.network_id
			WHERE biomass_aggregator.archived_at IS NULL
				AND (true=$2 OR biomass_aggregator.name ILIKE '%%' || $1 || '%%')`
	err := r.Get(&count, SQL, args...)
	if err != nil {
		return 0, errors.Wrap(err, "getBiomassAggregatorCount")
	}
	return count, nil
}

// getBiomassAggregatorById get biomass aggregator by id
func (r *Repository) getBiomassAggregatorById(biomassAggregatorUrlID string) (biomassAggregator biomassAggregatorDetailsResponse, err error) {
	SQL := `SELECT biomass_aggregator.id,
       			   biomass_aggregator.name,
       			   users.email,
       			   biomass_aggregator.trained,
       			   biomass_aggregator.location_name,
       			   biomass_aggregator.location,
       			   users.number,
       			   users.country_code
			FROM users
			    join biomass_aggregator_manager on users.id = biomass_aggregator_manager.manager_id
				join biomass_aggregator on biomass_aggregator_manager.biomass_aggregator_id = biomass_aggregator.id
			WHERE biomass_aggregator.id = $1
			AND biomass_aggregator.archived_at IS NULL`
	err = r.Get(&biomassAggregator, SQL, biomassAggregatorUrlID)
	if err != nil {
		return biomassAggregator, errors.Wrap(err, "getBiomassAggregatorById")
	}
	return biomassAggregator, nil
}

// editBiomassAggregator edit biomass aggregator
func (r *Repository) editBiomassAggregator(biomassAggregatorUrlID string, biomassAggregator biomassAggregatorDetailsRequest) error {
	var userID string
	location := fmt.Sprintf("(%v,%v)", biomassAggregator.Latitude, biomassAggregator.Longitude)

	SQL := `SELECT biomass_aggregator_manager.manager_id 
			FROM biomass_aggregator_manager
			WHERE biomass_aggregator_manager.biomass_aggregator_id = $1
			and biomass_aggregator_manager.archived_at is null`
	err := r.Get(&userID, SQL, biomassAggregatorUrlID)
	if err != nil {
		return errors.Wrap(err, "editBiomassAggregator")
	}

	args := []interface{}{
		userID,
		biomassAggregator.Name,
	}
	SQL = `UPDATE users
			SET name = $2
			WHERE id = $1
			AND archived_at IS NULL`
	_, err = r.Exec(SQL, args...)
	if err != nil {
		return errors.Wrap(err, "editBiomassAggregator")
	}

	args = []interface{}{
		biomassAggregatorUrlID,
		biomassAggregator.Name,
		biomassAggregator.Trained,
		biomassAggregator.LocationName,
		location,
	}
	SQL = `UPDATE biomass_aggregator
			SET name = $2,
			    trained =$3,
				location_name = $4,
				location = $5
			WHERE id = $1
			AND archived_at IS NULL`
	_, err = r.Exec(SQL, args...)
	if err != nil {
		return errors.Wrap(err, "editBiomassAggregator")
	}

	return nil
}

// deleteBiomassAggregator delete biomass aggregator
func (r *Repository) deleteBiomassAggregator(biomassAggregatorUrlID string) error {
	var userID string

	var farmerCount int
	SQL := `SELECT count(distinct farmer_details.user_id)
			FROM biomass_aggregator
				 left join network on biomass_aggregator.id = network.biomass_aggregator_id
				 left join farmer_details on network.id = farmer_details.network_id
			WHERE biomass_aggregator.id = $1
				AND biomass_aggregator.archived_at IS NULL
				GROUP BY biomass_aggregator.id`
	err := r.Get(&farmerCount, SQL, biomassAggregatorUrlID)
	if err != nil {
		return errors.Wrap(err, "deleteBiomassAggregator")
	}
	if farmerCount != 0 {
		return errors.New("deleteBiomassAggregator, farmer count is not zero")
	}

	SQL = `SELECT users.id 
			FROM users
			    join biomass_aggregator_manager on users.id = biomass_aggregator_manager.manager_id
				join biomass_aggregator on biomass_aggregator_manager.biomass_aggregator_id = biomass_aggregator.id
			WHERE biomass_aggregator.id = $1
			AND biomass_aggregator.archived_at IS NULL`
	err = r.Get(&userID, SQL, biomassAggregatorUrlID)
	if err != nil {
		return errors.Wrap(err, "deleteBiomassAggregator")
	}

	SQL = `UPDATE biomass_aggregator
			SET archived_at = now()
			WHERE id = $1
			AND archived_at IS NULL`
	_, err = r.Exec(SQL, biomassAggregatorUrlID)
	if err != nil {
		return errors.Wrap(err, "deleteBiomassAggregator")
	}

	SQL = `UPDATE biomass_aggregator_manager
			SET archived_at = now()
			WHERE biomass_aggregator_id = $1
			AND archived_at IS NULL`
	_, err = r.Exec(SQL, biomassAggregatorUrlID)
	if err != nil {
		return errors.Wrap(err, "deleteBiomassAggregator")
	}

	SQL = `UPDATE users
			SET archived_at = now()
			WHERE id = $1
			AND archived_at IS NULL`
	_, err = r.Exec(SQL, userID)
	if err != nil {
		return errors.Wrap(err, "deleteBiomassAggregator")
	}

	return nil
}

// getBANetworkList get BA network list
func (r *Repository) getBANetworkList(filters models.GenericFilters, biomassAggregatorID string) ([]bANetwork, error) {
	args := []interface{}{
		filters.Limit,
		filters.Limit * filters.Page,
		filters.SearchString,
		filters.SearchString == "",
		biomassAggregatorID,
	}
	bANetworks := make([]bANetwork, 0)
	SQL := `SELECT network.id,
				   network.name,
				   network.location_name,
				   count(distinct farmer_details.user_id) AS farmers_count,
				   users.name As manager_name
			FROM network
				 left join network_manager on network.id = network_manager.network_id
			     left join users on network_manager.manager_id = users.id
				 left join farmer_details on network.id = farmer_details.network_id
			WHERE network.archived_at IS NULL AND 
			  	  network.biomass_aggregator_id = $5
				AND (true=$4 OR network.name ILIKE '%%' || $3 || '%%')
			GROUP BY network.id, network.created_at, users.name
			ORDER BY network.created_at DESC
			LIMIT $1 OFFSET $2`
	err := r.Select(&bANetworks, SQL, args...)
	if err != nil {
		return bANetworks, errors.Wrap(err, "getBiomassAggregator")
	}
	return bANetworks, nil
}

// getBANetworkListCount get BA network list count
func (r *Repository) getBANetworkListCount(filters models.GenericFilters, biomassAggregatorID string) (int, error) {
	var count int
	args := []interface{}{
		filters.SearchString,
		filters.SearchString == "",
		biomassAggregatorID,
	}
	SQL := `SELECT COUNT(distinct network.id)
			FROM network
				 left join network_manager on network.id = network_manager.network_id
			     left join users on network_manager.manager_id = users.id
				 left join farmer_details on network.id = farmer_details.network_id
			WHERE network.archived_at IS NULL AND
			     network.biomass_aggregator_id = $3
				AND (true=$2 OR network.name ILIKE '%%' || $1 || '%%')`
	err := r.Get(&count, SQL, args...)
	if err != nil {
		return 0, errors.Wrap(err, "getBiomassAggregatorCount")
	}
	return count, nil
}

// createCSNetwork create C S network
func (r *Repository) createCSNetwork(network networkRequest) error {
	var networkId string

	args := []interface{}{
		network.Name,
		network.LocationName,
		network.BAggregatorID,
	}

	SQL := `INSERT INTO network (
                 name,
    			 location_name,
                 biomass_aggregator_id
    		)
			VALUES ($1,$2,$3) RETURNING id`
	err := r.Get(&networkId, SQL, args...)
	if err != nil {
		return errors.Wrap(err, "createCSNetwork")
	}

	return nil
}

// getCSNetwork get C S network
func (r *Repository) getCSNetwork(filters models.GenericFilters) ([]network, error) {

	args := []interface{}{
		filters.Limit,
		filters.Limit * filters.Page,
		filters.SearchString,
		filters.SearchString == "",
	}

	networks := make([]network, 0)

	SQL := `SELECT network.id,
				   network.name,
				   network.location_name,
				   count(distinct farmer_details.user_id) AS farmers_count,
				   network_manager.id AS manager_id,
				   users.name AS manager_name,
				   users.email AS manager_email,
				   users.number AS manager_number,
				   users.country_code AS manager_country_code,
				   count(distinct k.id) AS kiln_count,
				   biomass_aggregator.id AS biomass_aggregator_id,
				   biomass_aggregator.name AS biomass_aggregator_name
			FROM network
				 left join network_manager on network.id = network_manager.network_id
			     left join users on network_manager.manager_id = users.id
				 left join farmer_details on network.id = farmer_details.network_id
				 left join biomass_aggregator on network.biomass_aggregator_id = biomass_aggregator.id
			     left join klin k on network.id = k.network_id
			WHERE network.archived_at IS NULL
				AND (true=$4 OR network.name ILIKE '%%' || $3 || '%%')
			GROUP BY network.id, network.created_at, network_manager.id, users.id, biomass_aggregator.id
			ORDER BY network.created_at DESC
			LIMIT $1 OFFSET $2`
	err := r.Select(&networks, SQL, args...)
	if err != nil {
		return networks, errors.Wrap(err, "getCSNetwork")
	}
	return networks, nil
}

// getCSNetworkCount get C S network count
func (r *Repository) getCSNetworkCount(filters models.GenericFilters) (int, error) {
	var count int
	args := []interface{}{
		filters.SearchString,
		filters.SearchString == "",
	}
	SQL := `SELECT COUNT(distinct network.id)
			FROM network
				 left join network_manager on network.id = network_manager.network_id
			     left join users on network_manager.manager_id = users.id
				 left join farmer_details on network.id = farmer_details.network_id
				 left join biomass_aggregator on network.biomass_aggregator_id = biomass_aggregator.id
				 left join klin k on network.id = k.network_id
			WHERE network.archived_at IS NULL
				AND (true=$2 OR network.name ILIKE '%%' || $1 || '%%')`
	err := r.Get(&count, SQL, args...)
	if err != nil {
		return 0, errors.Wrap(err, "getCSNetworkCount")
	}
	return count, nil
}

// editCSNetwork edit C S network
func (r *Repository) editCSNetwork(id string, network editNetworkRequest) error {
	var networkId string
	var userID string

	SQL := `SELECT id 
			FROM network
			WHERE id = $1
			AND archived_at IS NULL`
	err := r.Get(&networkId, SQL, id)
	if err != nil {
		return errors.Wrap(err, "editCSNetwork")
	}

	args := []interface{}{
		network.Name,
		network.LocationName,
		network.BAggregatorID,
		id,
	}
	SQL = `UPDATE network
			SET name = $1,
				location_name =$2,
				biomass_aggregator_id = $3
			WHERE id = $4
			AND archived_at IS NULL`
	_, err = r.Exec(SQL, args...)
	if err != nil {
		return errors.Wrap(err, "editCSNetwork")
	}

	args = []interface{}{
		id,
		time.Now(),
	}
	SQL = `UPDATE network_manager
			SET archived_at = $2
			WHERE network_id = $1
			AND archived_at IS NULL`
	_, err = r.Exec(SQL, args...)
	if err != nil {
		return errors.Wrap(err, "editCSNetwork")
	}

	SQL = `SELECT manager_id 
			FROM network_manager 
			WHERE id = $1`
	err = r.Get(&userID, SQL, network.CSMangerID)
	if err != nil {
		return errors.Wrap(err, "editCSNetwork")
	}

	args = []interface{}{
		networkId,
		userID,
	}
	SQL = `INSERT INTO network_manager (
            network_id, 
            manager_id
           ) 
           VALUES ($1,$2)`
	_, err = r.Exec(SQL, args...)
	if err != nil {
		return errors.Wrap(err, "editCSNetwork")
	}

	return nil
}

// deleteCSNetwork delete c s network
func (r *Repository) deleteCSNetwork(id string) error {

	SQL := `UPDATE network
			SET archived_at = now()
			WHERE id = $1
			AND archived_at IS NULL`
	_, err := r.Exec(SQL, id)
	if err != nil {
		return errors.Wrap(err, "deleteCSNetwork")
	}

	SQL = `UPDATE network_manager
			SET archived_at = $2
			WHERE network_id = $1
			AND archived_at IS NULL`
	_, err = r.Exec(SQL, id, time.Now())
	if err != nil {
		return errors.Wrap(err, "deleteCSNetwork")
	}
	return nil
}

// createCSNetworkManager create C S network manager
func (r *Repository) createCSNetworkManager(networkID string, managerID string) error {

	SQL := `INSERT INTO network_manager (
    			 manager_id,
                 network_id
    		)
			VALUES ($1, $2)`
	_, err := r.Exec(SQL, managerID, networkID)
	if err != nil {
		return errors.Wrap(err, "createCSNetworkManager")
	}

	return nil
}

// getBAApprovedFarmerList get B A Approved Farmer List
func (r *Repository) getBAApprovedFarmerList(filters models.GenericFilters, biomassAggregatorID string) ([]bAFarmer, error) {
	args := []interface{}{
		filters.Limit,
		filters.Limit * filters.Page,
		filters.SearchString,
		filters.SearchString == "",
		biomassAggregatorID,
	}
	bAFarmers := make([]bAFarmer, 0)
	SQL := `SELECT users.id,
				   users.name,
				   users.address,
				   users.number,
				   users.country_code,
				   count(distinct farms.id) AS farms_count,
				   network.id AS network_id,
				   network.name AS network_name
			FROM users
			     left join farms on users.id = farms.farmer_id
			     left join farmer_details on users.id = farmer_details.user_id
				 left join network on network.id = farmer_details.network_id
			WHERE users.archived_at IS NULL 
			  AND network.biomass_aggregator_id = $5
			  AND (true=$4 OR network.name ILIKE '%%' || $3 || '%%')
			GROUP BY users.id, users.created_at, network.id
			ORDER BY users.created_at DESC
			LIMIT $1 OFFSET $2`
	err := r.Select(&bAFarmers, SQL, args...)
	if err != nil {
		return bAFarmers, errors.Wrap(err, "getBiomassAggregator")
	}
	return bAFarmers, nil
}

// getBAApprovedFarmerListCount getBAApprovedFarmerListCount
func (r *Repository) getBAApprovedFarmerListCount(filters models.GenericFilters, biomassAggregatorID string) (int, error) {
	var count int
	args := []interface{}{
		filters.SearchString,
		filters.SearchString == "",
		biomassAggregatorID,
	}
	SQL := `SELECT count(users.id)
			FROM users
			     left join farms on users.id = farms.farmer_id
			     left join farmer_details on users.id = farmer_details.user_id
				 left join network on network.id = farmer_details.network_id
			WHERE users.archived_at IS NULL 
			  AND network.biomass_aggregator_id = $3
			  AND (true=$2 OR network.name ILIKE '%%' || $1 || '%%')`
	err := r.Get(&count, SQL, args...)
	if err != nil {
		return 0, errors.Wrap(err, "getBiomassAggregatorCount")
	}
	return count, nil
}

// getBAPendingFarmerList getBAPendingFarmerList
func (r *Repository) getBAPendingFarmerList(filters models.GenericFilters, biomassAggregatorId string) ([]bAFarmer, error) {
	args := []interface{}{
		filters.Limit,
		filters.Limit * filters.Page,
		models.UserAccountTypeFarmer,
		biomassAggregatorId,
	}
	bAFarmers := make([]bAFarmer, 0)
	SQL := `SELECT users.id,
				   users.name,
				   users.address,
				   users.number,
				   users.country_code,
				   count(distinct farms.id) AS farms_count
		FROM farmer_details
				 left join farms on farmer_details.user_id = farms.farmer_id
				 left join users on users.id = farmer_details.user_id
				 left join farmer_rejected on users.id = farmer_rejected.farmer_id 
				                                  		 AND
                                                         farmer_rejected.biomass_aggregator_id = $4
			WHERE users.archived_at IS NULL
			  AND users.account_type = $3
			  AND farmer_details.network_id IS NULL
			  AND farmer_rejected.farmer_id IS NULL 
			GROUP BY users.id, users.created_at
			ORDER BY users.created_at DESC
			LIMIT $1 OFFSET $2`
	err := r.Select(&bAFarmers, SQL, args...)
	if err != nil {
		return bAFarmers, errors.Wrap(err, "getBiomassAggregator")
	}
	return bAFarmers, nil
}

// getBAPendingFarmerListCount getBAPendingFarmerListCount
func (r *Repository) getBAPendingFarmerListCount(filters models.GenericFilters, id string) (int, error) {
	var count int
	SQL := `SELECT count(distinct users.id)
			FROM farmer_details
         		left join farms on farmer_details.user_id = farms.farmer_id
         		left join users on users.id = farmer_details.user_id
				left join farmer_rejected on users.id = farmer_rejected.farmer_id
							                            AND
                                                        farmer_rejected.biomass_aggregator_id = $2
			WHERE users.archived_at IS NULL
			  AND users.account_type = $1
			  AND farmer_details.network_id IS NULL
			  AND farmer_rejected.farmer_id IS NULL`
	err := r.Get(&count, SQL, models.UserAccountTypeFarmer, id)
	if err != nil {
		return 0, errors.Wrap(err, "getBiomassAggregatorCount")
	}
	return count, nil
}

// createKiln create kiln
func (r *Repository) createKiln(kiln kilnRequest) error {

	args := []interface{}{
		kiln.Name,
		kiln.Address,
		kiln.NetworkID,
	}

	SQL := `INSERT INTO klin (
                 name,
    			 address,
                 network_id
    		)
			VALUES ($1,$2,$3)`
	_, err := r.Exec(SQL, args...)
	if err != nil {
		return errors.Wrap(err, "createKiln")
	}

	return nil
}

// getKiln get get kiln
func (r *Repository) getKiln(filters models.GenericFilters) ([]kiln, error) {

	args := []interface{}{
		filters.Limit,
		filters.Limit * filters.Page,
		filters.SearchString,
		filters.SearchString == "",
	}

	kilns := make([]kiln, 0)

	SQL := `SELECT klin.id,
				   klin.name,
				   klin.address,
				   array_remove(ARRAY_AGG(klin_operator.id), NUll) AS kiln_operator_ids,
				   network.id AS network_id,
				   network.name AS network_name 
			FROM klin
				 left join klin_operator on klin.id = klin_operator.klin_id
			     left join users on klin_operator.operator_id = users.id
			     left join network on klin.network_id = network.id
			WHERE klin.archived_at IS NULL
				AND (true=$4 OR klin.name ILIKE '%%' || $3 || '%%')
			GROUP BY klin.id, klin.created_at, network.id
			ORDER BY klin.created_at DESC
			LIMIT $1 OFFSET $2`
	err := r.Select(&kilns, SQL, args...)
	if err != nil {
		return kilns, errors.Wrap(err, "getCSNetwork")
	}
	return kilns, nil
}

// getKilnCount get kiln count
func (r *Repository) getKilnCount(filters models.GenericFilters) (int, error) {
	var count int
	args := []interface{}{
		filters.SearchString,
		filters.SearchString == "",
	}
	SQL := `SELECT COUNT(distinct klin.id)
			FROM klin
				 left join klin_operator on klin.id = klin_operator.klin_id
			     left join users on klin_operator.operator_id = users.id
				 left join network on klin.network_id = network.id
			WHERE klin.archived_at IS NULL
				AND (true=$2 OR klin.name ILIKE '%%' || $1 || '%%')`
	err := r.Get(&count, SQL, args...)
	if err != nil {
		return 0, errors.Wrap(err, "getCSNetworkCount")
	}
	return count, nil
}

// editKiln edit kiln
func (r *Repository) editKiln(id string, kiln editKilnRequest) error {
	var kilnId string

	SQL := `SELECT id 
			FROM klin
			WHERE id = $1
			AND archived_at IS NULL`
	err := r.Get(&kilnId, SQL, id)
	if err != nil {
		return errors.Wrap(err, "editKiln")
	}

	args := []interface{}{
		kiln.Name,
		kiln.Address,
		id,
	}
	SQL = `UPDATE klin
			SET name = $1,
				address =$2
			WHERE id = $3
			AND archived_at IS NULL`
	_, err = r.Exec(SQL, args...)
	if err != nil {
		return errors.Wrap(err, "editKiln")
	}

	return nil
}

// deleteKiln delete kiln
func (r *Repository) deleteKiln(id string) error {

	SQL := `UPDATE klin
			SET archived_at = now()
			WHERE id = $1
			AND archived_at IS NULL`
	_, err := r.Exec(SQL, id)
	if err != nil {
		return errors.Wrap(err, "deleteKiln")
	}
	return nil
}

// createKilnOperator create kiln operator
func (r *Repository) createKilnOperator(kilnOperator kilnOperatorRequest) error {

	SQL := `INSERT INTO klin_operator (
    			 operator_id,
                 klin_id
    		)
			VALUES ($1, $2)`
	_, err := r.Exec(SQL, kilnOperator.UserID, kilnOperator.KilnID)
	if err != nil {
		return errors.Wrap(err, "createKilnOperator")
	}

	return nil
}

// fetchKilnOperators fetch kiln operators
func (r *Repository) fetchKilnOperators(kilnOperatorsIds pq.StringArray) ([]kilnOperator, error) {
	kilnOperators := make([]kilnOperator, 0)

	SQL := `SELECT klin_operator.id,
				   users.name,
				   users.number,
				   users.country_code,
				   klin_operator.klin_id AS klin_id
			FROM klin_operator 
				JOIN users on users.id = klin_operator.operator_id
			WHERE klin_operator.id = ANY ($1)`

	err := r.Select(&kilnOperators, SQL, kilnOperatorsIds)
	if err != nil {
		return nil, errors.Wrap(err, "fetchKilnOperators")
	}

	return kilnOperators, nil
}

// assigningFarmerToCSNetwork assigning Farmer To C S Network
func (r *Repository) assigningFarmerToCSNetwork(farmer farmerId, id string) error {
	var networkId string

	SQL := `SELECT id 
			FROM network
			WHERE id = $1
			AND archived_at IS NULL`
	err := r.Get(&networkId, SQL, id)
	if err != nil {
		return errors.Wrap(err, "assigningFarmerToCSNetwork")
	}

	args := []interface{}{
		id,
		farmer.FarmerID,
	}
	SQL = `UPDATE farmer_details
			SET network_id = $1
			WHERE user_id = $2`
	_, err = r.Exec(SQL, args...)
	if err != nil {
		return errors.Wrap(err, "assigningFarmerToCSNetwork")
	}

	return nil
}

// getBARejectedFarmerList getBARejectedFarmerList
func (r *Repository) getBARejectedFarmerList(filters models.GenericFilters, id string) ([]bAFarmer, error) {
	args := []interface{}{
		filters.Limit,
		filters.Limit * filters.Page,
		id,
	}
	bAFarmers := make([]bAFarmer, 0)
	SQL := `SELECT users.id,
				   users.name,
				   users.address,
				   users.number,
				   users.country_code
			FROM users
			     left join farmer_rejected fr on users.id = fr.farmer_id
			     join biomass_aggregator on fr.biomass_aggregator_id = biomass_aggregator.id
			WHERE users.archived_at IS NULL
			   AND fr.biomass_aggregator_id = $3
			GROUP BY users.id, users.created_at
			ORDER BY users.created_at DESC
			LIMIT $1 OFFSET $2`
	err := r.Select(&bAFarmers, SQL, args...)
	if err != nil {
		return bAFarmers, errors.Wrap(err, "getBARejectedFarmerList")
	}
	return bAFarmers, nil
}

// getBARejectedFarmerListCount get B A Rejected Farmer List
func (r *Repository) getBARejectedFarmerListCount(filters models.GenericFilters, id string) (int, error) {
	var count int
	args := []interface{}{
		id,
	}
	SQL := `SELECT count(users.id)
			FROM users
			     left join farmer_rejected fr on users.id = fr.farmer_id
			     join biomass_aggregator on fr.biomass_aggregator_id = biomass_aggregator.id
			WHERE users.archived_at IS NULL
			   AND fr.biomass_aggregator_id = $1`
	err := r.Get(&count, SQL, args...)
	if err != nil {
		return 0, errors.Wrap(err, "getBARejectedFarmerListCount")
	}
	return count, nil
}

// getCSNKilnList  get C S N by id kiln list
func (r *Repository) getCSNKilnList(filters models.GenericFilters, csnID string) ([]kiln, error) {

	args := []interface{}{
		filters.Limit,
		filters.Limit * filters.Page,
		filters.SearchString,
		filters.SearchString == "",
		csnID,
	}

	kilns := make([]kiln, 0)

	SQL := `SELECT klin.id,
				   klin.name,
				   klin.address,
				   array_remove(ARRAY_AGG(klin_operator.id), NUll) AS kiln_operator_ids
			FROM klin
				 left join klin_operator on klin.id = klin_operator.klin_id
			     left join users on klin_operator.operator_id = users.id
				 left join network on network.id = klin.network_id
			WHERE klin.archived_at IS NULL
				AND (true=$4 OR klin.name ILIKE '%%' || $3 || '%%')
				AND network_id = $5
			GROUP BY klin.id, klin.created_at
			ORDER BY klin.created_at DESC
			LIMIT $1 OFFSET $2`
	err := r.Select(&kilns, SQL, args...)
	if err != nil {
		return kilns, errors.Wrap(err, "getCSNKilnList")
	}
	return kilns, nil
}

// getCSNKilnListCount get C S N by id kiln list count
func (r *Repository) getCSNKilnListCount(filters models.GenericFilters, csnID string) (int, error) {
	var count int
	args := []interface{}{
		filters.SearchString,
		filters.SearchString == "",
		csnID,
	}
	SQL := `SELECT COUNT(distinct klin.id)
			FROM klin
				 left join klin_operator on klin.id = klin_operator.klin_id
			     left join users on klin_operator.operator_id = users.id
			     left join network on network.id = klin.network_id
			WHERE klin.archived_at IS NULL
				AND (true=$2 OR klin.name ILIKE '%%' || $1 || '%%')
				AND network_id = $3`
	err := r.Get(&count, SQL, args...)
	if err != nil {
		return 0, errors.Wrap(err, "getCSNKilnListCount")
	}
	return count, nil
}

// getCSNFarmerList get C S N by id Farmer List
func (r *Repository) getCSNFarmerList(filters models.GenericFilters, csnID string) ([]bAFarmer, error) {
	args := []interface{}{
		filters.Limit,
		filters.Limit * filters.Page,
		filters.SearchString,
		filters.SearchString == "",
		csnID,
	}
	bAFarmers := make([]bAFarmer, 0)
	SQL := `SELECT users.id,
				   users.name,
				   users.address,
				   users.number,
				   users.country_code,
				   count(distinct farms.id) AS farms_count
			FROM users
			     left join farms on users.id = farms.farmer_id
			     left join farmer_details on users.id = farmer_details.user_id
				 left join network on network.id = farmer_details.network_id
			WHERE users.archived_at IS NULL 
			  AND (true=$4 OR network.name ILIKE '%%' || $3 || '%%')
			  AND network.id = $5
			GROUP BY users.id, users.created_at
			ORDER BY users.created_at DESC
			LIMIT $1 OFFSET $2`
	err := r.Select(&bAFarmers, SQL, args...)
	if err != nil {
		return bAFarmers, errors.Wrap(err, "getCSNFarmerList")
	}
	return bAFarmers, nil
}

// getCSNFarmerListCount get C S N by id farmer list count
func (r *Repository) getCSNFarmerListCount(filters models.GenericFilters, csnID string) (int, error) {
	var count int
	args := []interface{}{
		filters.SearchString,
		filters.SearchString == "",
		csnID,
	}
	SQL := `SELECT count(distinct users.id)
			FROM users
			     left join farms on users.id = farms.farmer_id
			     left join farmer_details on users.id = farmer_details.user_id
				 left join network on network.id = farmer_details.network_id
			WHERE users.archived_at IS NULL 
			  AND network.id = $3
			  AND (true=$2 OR network.name ILIKE '%%' || $1 || '%%')`
	err := r.Get(&count, SQL, args...)
	if err != nil {
		return 0, errors.Wrap(err, "getCSNFarmerListCount")
	}
	return count, nil
}

// bARejectFarmer B A reject farmer
func (r *Repository) bARejectFarmer(farmer farmerId, id string) error {
	var biomassAggregatorId string

	SQL := `SELECT id 
			FROM biomass_aggregator
			WHERE id = $1
			AND archived_at IS NULL`
	err := r.Get(&biomassAggregatorId, SQL, id)
	if err != nil {
		return errors.Wrap(err, "bARejectFarmer")
	}

	args := []interface{}{
		id,
		farmer.FarmerID,
	}
	SQL = `INSERT INTO farmer_rejected (
			biomass_aggregator_id,
            farmer_id
			) 
			VALUES ($1, $2)`
	_, err = r.Exec(SQL, args...)
	if err != nil {
		return errors.Wrap(err, "bARejectFarmer")
	}

	return nil
}

// getCSNetworkDetailsByID get C S Network Details By ID
func (r *Repository) getCSNetworkDetailsByID(id string) (network, error) {

	var csn network

	SQL := `SELECT network.id,
				   network.name,
				   network.location_name,
				   count(distinct farmer_details.user_id) AS farmers_count,
				   network_manager.id AS manager_id,
				   users.name AS manager_name,
				   users.email AS manager_email,
				   biomass_aggregator.id AS biomass_aggregator_id,
				   biomass_aggregator.name AS biomass_aggregator_name
			FROM network
				 left join network_manager on network.id = network_manager.network_id
			     left join users on network_manager.manager_id = users.id
				 left join farmer_details on network.id = farmer_details.network_id
				 left join biomass_aggregator on network.biomass_aggregator_id = biomass_aggregator.id
			WHERE network.archived_at IS NULL 
			     AND network.id = $1
			GROUP BY network.id, network.created_at, network_manager.id, users.id, biomass_aggregator.id`
	err := r.Get(&csn, SQL, id)
	if err != nil {
		return csn, errors.Wrap(err, "getCSNetworkDetailsByID")
	}
	return csn, nil
}
