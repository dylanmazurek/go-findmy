package models

type FcmSession struct {
	RegistrationToken     *string  `json:"registrationToken"`
	PrivateKeyBase64      *string  `json:"privateKey"`
	AuthSecret            *string  `json:"authSecret"`
	InstallationAuthToken *string  `json:"installationAuthToken"`
	PersistentIds         []string `json:"persistentIds"`
}
