package relution

import (
	"encoding/json"
	"fmt"
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
	stmtIns, err := r.database.DB.Prepare("INSERT INTO devices VALUES( ?, ?, ?, ?, ?, ?, ? ) ON DUPLICATE KEY UPDATE id = ?, name = ?, loggedin_user = ?, device_type = ?, battery_level = ?, device_group = ?, device_group_index = ?")
	if err != nil {
		log.Errorf("Error preparing insert statement: %v", err)
		os.Exit(1)
	}
	//noinspection GoUnhandledErrorResult
	defer stmtIns.Close()

	for _, rDevice := range devicesResponse.Results {
		gDevice, err := models.RelutionDeviceToGeneralDevice(rDevice)
		if err != nil {
			log.Warnf("Error converting relution device to general device: %v", err)
			continue
		}
		_, err = stmtIns.Exec(gDevice.Id, gDevice.Name, gDevice.LoggedinUser, gDevice.DeviceType, gDevice.BatteryLevel, gDevice.DeviceGroup, gDevice.DeviceGroupIndex, gDevice.Id, gDevice.Name, gDevice.LoggedinUser, gDevice.DeviceType, gDevice.BatteryLevel, gDevice.DeviceGroup, gDevice.DeviceGroupIndex)
		if err != nil {
			log.Warnf("Error executing insert statement: %v", err)
		}
	}
	log.Infof("Fetched devices...")
}
