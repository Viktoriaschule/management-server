package relution

import (
	"encoding/json"
	"fmt"
	"github.com/viktoriaschule/management-server/charging"
	"github.com/viktoriaschule/management-server/helper"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/viktoriaschule/management-server/config"
	"github.com/viktoriaschule/management-server/database"
	"github.com/viktoriaschule/management-server/log"
	"github.com/viktoriaschule/management-server/models"
)

type Relution struct {
	config   *config.Config
	database *database.Database
	devices  []models.RelutionDevice
}

func NewRelution(config *config.Config, database *database.Database) *Relution {
	return &Relution{config: config, database: database}
}

func (r *Relution) FetchDevices() {
	log.Infof("Fetching devices...")
	url := fmt.Sprintf("https://%s/relution/api/v1/devices", r.config.Relution.Host)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("Error reading request: %v", err)
		os.Exit(1)
	}

	req.Header.Set("X-User-Access-Token", r.config.Relution.Token)

	client := &http.Client{Timeout: time.Second * 10}

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Error reading response: %v", err)
		os.Exit(1)
	}
	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading body: %v", err)
		os.Exit(1)
	}

	var devicesResponse relutionDevicesResponse
	err = json.Unmarshal(body, &devicesResponse)
	if err != nil {
		log.Errorf("Error parsing json: %v", err)
		os.Exit(1)
	}
	stmtIns, err := r.database.DB.Prepare("INSERT INTO devices VALUES( ?, ?, ?, ?, ?, ?, ?, ? ) ON DUPLICATE KEY UPDATE id = ?, name = ?, loggedin_user = ?, device_type = ?, battery_level = ?, is_charging = ?, device_group = ?, device_group_index = ?")
	if err != nil {
		log.Errorf("Error preparing insert statement: %v", err)
		os.Exit(1)
	}
	//noinspection GoUnhandledErrorResult
	defer stmtIns.Close()

	// Get all current devices
	_oldDevices, err := getLoadedDevices(r.database, "")
	if err != nil {
		log.Errorf("Error fetching old devices: %v", err)
		os.Exit(1)
	}

	// Convert devices list to map
	oldDevices := make(map[string]models.GeneralDevice)
	for _, device := range *_oldDevices {
		oldDevices[device.Id] = device
	}

	changedCount := 0

	// Start charging sync
	charging.StartSync(r.database)

	for _, rDevice := range devicesResponse.Results {
		// Get the device
		gDevice, err := models.RelutionDeviceToGeneralDevice(rDevice)
		if err != nil {
			log.Warnf("Error converting relution device to general device: %v", err)
			continue
		}

		// Add the battery entry if changed
		oldDevice, isOld := oldDevices[gDevice.Id]

		// Sync the charging mode for the device
		charging.SyncDevice(gDevice, &oldDevice, !isOld)

		// Add or change device entry
		if !isOld || models.HasDeviceChanged(gDevice, &oldDevice) {
			_, err = stmtIns.Exec(gDevice.Id, gDevice.Name, gDevice.LoggedinUser, gDevice.DeviceType, gDevice.BatteryLevel, gDevice.IsCharging, gDevice.DeviceGroup, gDevice.DeviceGroupIndex, gDevice.Id, gDevice.Name, gDevice.LoggedinUser, gDevice.DeviceType, gDevice.BatteryLevel, gDevice.IsCharging, gDevice.DeviceGroup, gDevice.DeviceGroupIndex)
			if err != nil {
				log.Warnf("Error executing insert statement: %v", err)
			}
			changedCount++
		}
	}
	log.Infof("Fetched devices (%d have changed)", changedCount)

	charging.EndSync(r.database)
}

func GetValidLoadedDevices(database *database.Database) (devices *[]models.GeneralDevice, err error) {
	return getLoadedDevices(database, "WHERE device_group != 0 OR device_type = 1")
}

func getLoadedDevices(database *database.Database, filter string) (devices *[]models.GeneralDevice, err error) {
	rows, _err := database.DB.Query("SELECT * FROM devices" + " " + filter)
	if _err != nil {
		err = &helper.LoadError{Msg: fmt.Sprintf("Database query failed %v", _err)}
		return nil, err
	}
	var _devices []models.GeneralDevice
	device := &models.GeneralDevice{}
	//noinspection GoUnhandledErrorResult
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&device.Id, &device.Name, &device.LoggedinUser, &device.DeviceType, &device.BatteryLevel, &device.IsCharging, &device.DeviceGroup, &device.DeviceGroupIndex)
		if err != nil {
			err = &helper.LoadError{Msg: "Database query failed"}
			return nil, err
		}
		_devices = append(_devices, *device)
	}
	err = rows.Err()
	if err != nil {
		err = &helper.LoadError{Msg: "Database query failed"}
		return nil, err
	}
	devices = &_devices

	return devices, err
}
