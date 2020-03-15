package models

import (
	"regexp"
	"strconv"
	"strings"
)

type GeneralDevice struct {
	Id               string  `json:"id"`
	Name             string  `json:"name"`
	LoggedinUser     string  `json:"loggedin_user"`
	DeviceType       int64   `json:"device_type"`
	BatteryLevel     float64 `json:"battery_level"`
	DeviceGroup      int64   `json:"device_group"`
	DeviceGroupIndex string  `json:"device_group_index"`
}

type RelutionDevice struct {
	Uuid               string
	OrganizationUuid   string
	User               string
	Username           string
	Name               string
	Status             string
	Manufacturer       string
	Model              string
	Platform           string
	EnrollmentType     string
	Os                 string
	Ownership          string
	PushUdid           string
	PushToken          string
	PushMagic          string
	AppPushToken       string
	LastConnectionDate int
	EnrollmentDate     int
	ModificationDate   int
	UnlockToken        string
	Migrated           bool
	Jailbroken         bool
	PushEnabled        bool
	BoundApp           string
	ExecutedPolicy     struct {
		PolicyUuid        string
		PolicyVersionUuid string
		Name              string
		Version           int
		State             string
	}
	InstalledApps []struct {
		DeviceUuid                 string
		Name                       string
		Identifier                 string
		Version                    string
		ShortVersion               string
		Managed                    bool
		IsValidated                bool
		BundleSize                 int
		DynamicSize                int
		Installing                 bool
		AppStoreVendable           bool
		DeviceBasedVpp             bool
		BetaApp                    bool
		AdHocCodeSigned            bool
		HasUpdateAvailable         bool
		MetaData                   []struct{} //TODO: check what the correct type is
		ManagedAppStatus           string
		ManagedAppHasConfiguration bool
		Validated                  bool
	}
	ManagedApps []struct {
		Uuid                      string
		DeviceUuid                string
		Identifier                string
		ExternalVersionIdentifier string
		HasConfiguration          bool
		HasFeedback               bool
		IsValidated               bool
		ManagementFlags           int
		Status                    string
	}
	Details struct {
		Platform                      string
		Name                          string
		OsVersion                     string
		BuildVersion                  string
		ModelName                     string
		Model                         string
		ProductName                   string
		SerialNumber                  string
		DeviceCapacity                float64
		AvailableDeviceCapacity       float64
		BatteryLevel                  float64
		CellularTechnology            string
		IsSupervised                  bool
		IsDeviceLocatorServiceEnabled bool
		IsActivationLockEnabled       bool
		IsDoNotDisturbInEffect        bool
		IsCloudBackupEnabled          bool
		IsMDMLostModeEnabled          bool
		DeviceID                      string
		EasDeviceIdentifier           string
		BluetoothMAC                  string
		WifiMAC                       string
		EthernetMACs                  []string
		VoiceRoamingEnabled           bool
		DataRoamingEnabled            bool
		IsRoaming                     bool
		Profiles                      []struct {
			Uuid       string
			Name       string
			Identifier string
		}
		AvailableUpdates []struct{} //TODO: check what the correct type is
		CurrentOSUpdate  struct {
			ProductKey              string
			UpdateStatus            string
			LastUpdateStatusTime    int
			DownloadPercentComplete float64
			ErrorChain              []struct{} //TODO: check what the correct type is
			Downloaded              bool
		}
		DepDeviceUuid string
		DepProfile    struct {
			Uuid                    string
			CreatedBy               string
			CreationDate            int
			ModifiedBy              string
			ModificationDate        int
			ProfileUuid             string
			ProfileName             string
			Url                     string
			AllowPairing            bool
			IsSupervised            bool
			IsMandatory             bool
			IsAwaitDeviceConfigured bool
			IsMultiUser             bool
			IsMdmRemovable          bool
			SupportPhoneNumber      string
			SupportEmailAddress     string
			AnchorCerts             []struct{} //TODO: check what the correct type is
			SupervisingHostCerts    []struct{} //TODO: check what the correct type is
			SkipSetupItems          []string
			Department              string
			AuthenticationRequired  bool
		}
		Location struct {
			Accuracy  float64
			Time      int
			Altitude  float64
			Speed     float64
			Course    float64
			Latitude  float64
			Longitude float64
			Current   bool
		}
		Properties          struct{}   //TODO: check what the correct type is
		Resources           []struct{} //TODO: check what the correct type is
		SharedDeviceStatus  string
		AdvertisingMessages []struct{} //TODO: check what the correct type is
		Tags                []struct{} //TODO: check what the correct type is
		SimSlots            []struct{} //TODO: check what the correct type is
		Policy              struct {
			PolicyUuid        string
			PolicyVersionUuid string
			Name              string
			Version           int
			State             string
		}
		EnrolledInAndroidManagement bool
		IotDevice                   bool
		Ruleset                     struct {
			RulesetUuid        string
			RulesetVersionUuid string
			Name               string
			Version            int
		}
		ComplianceNoticeCount     int
		ComplianceViolatedCount   int
		SecuredSharedDeviceStatus string
		Enrollment                bool
	}
}

func RelutionDeviceToGeneralDevice(device RelutionDevice) (*GeneralDevice, error) {
	var deviceType int64 = 0
	var group int64 = 0
	var groupIndex = ""
	if strings.HasPrefix(strings.ToLower(device.Name), "l") {
		deviceType += 1
	}
	if len(strings.Split(device.Name, "-")) == 2 {
		fullGroup := strings.Split(device.Name, "-")[1]
		r, _ := regexp.Compile("[0-9]+")
		var err error
		group, err = strconv.ParseInt(r.FindString(fullGroup), 10, 64)
		r, _ = regexp.Compile("[a-z]+")
		groupIndex = r.FindString(fullGroup)
		if groupIndex == "" || err != nil {
			group = 0
		}
	}
	username := device.Username
	if username == "AACHEN-VSA Device User" {
		username = ""
	}
	return &GeneralDevice{
		Id:               strings.ToLower(device.Uuid),
		Name:             strings.ToLower(device.Name),
		LoggedinUser:     username,
		DeviceType:       deviceType,
		BatteryLevel:     device.Details.BatteryLevel,
		DeviceGroup:      group,
		DeviceGroupIndex: groupIndex,
	}, nil
}