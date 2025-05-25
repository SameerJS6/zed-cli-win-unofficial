package registry

type RegistryConfig struct {
	// AppName is the display name of the application in the Windows registry
	AppName string
	// ExecutablePath is the full path to the zed.exe executable file
	ExecutablePath string
	// AppUserModelID is a unique identifier for the application, typically the same as the executable path
	AppUserModelId string
	// GenericMenuText is the text displayed in the context menu for opening files with Zed
	GenericMenuText string
	// FileExtensions is a list of file extensions that Zed can open
	FileExtensions []string
	// PerFileTypeDescTmpl is a template string for describing file types in the registry
	PerFileTypeDescriptionText string
}

func NewConfig(executablePath string, extensions []string) *RegistryConfig {
	return &RegistryConfig{
		AppName:                    "Zed",
		ExecutablePath:             executablePath,
		AppUserModelId:             executablePath,
		GenericMenuText:            "Open w&ith Zed",
		FileExtensions:             extensions,
		PerFileTypeDescriptionText: "%s Source File (Zed)",
	}
}
