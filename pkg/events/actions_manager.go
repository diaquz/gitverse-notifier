package events

type ActionsManager struct {
	mapping map[string]*ActionSettings
	defaultSetting ActionSettings
}

// Read actions/*.yml and create mapping repository => ActionSettings
// Maybe some king of Manager, or just global var here
//
//	Error if no default / error during parsing
func SetupActionsManager() (ActionsManager, error) {
	return ActionsManager{}, nil
}

// ActionSetting by repository or default
func (m *ActionsManager) ActionSettingsByRepository(repository string) *ActionSettings {
	return &m.defaultSetting
}
