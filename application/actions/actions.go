package actions

// SocialAuthConfig holds the Key and Secret for a social auth provider
type SocialAuthConfig struct{ Key, Secret string }

// Don't Modify outside of this file.
var socialAuthConfigs = map[string]SocialAuthConfig{}

// Don't Modify outside of this file.
var socialAuthOptions = []authOption{}

func init() {
	socialAuthConfigs = getSocialAuthConfigs()
	socialAuthOptions = getSocialAuthOptions(socialAuthConfigs)
}
