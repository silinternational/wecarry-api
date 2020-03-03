package actions

// SocialAuthConfig holds the Key and Secret for a social auth provider
type SocialAuthConfig struct{ Key, Secret, Tenant string }

// Don't Modify outside of this file.
var socialAuthConfigs = map[string]SocialAuthConfig{}

// Don't Modify outside of this file.
var socialAuthSelectors = []authSelector{}

func init() {
	socialAuthConfigs = getSocialAuthConfigs()
	socialAuthSelectors = getSocialAuthSelectors(socialAuthConfigs)
}
