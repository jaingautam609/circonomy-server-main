package emailprovider

import (
	"circonomy-server/providers"
)

func verifyEmailTemplate() (*providers.DynamicTemplate, error) {
	return &providers.DynamicTemplate{
		TemplateID:  "d-107d6dc2aac84a08a2a38d2bff46eb9a",
		Categories:  []string{"Verify Email"},
		DynamicData: make(map[string]interface{}),
	}, nil
}

func forgotPasswordTemplate() (*providers.DynamicTemplate, error) {
	return &providers.DynamicTemplate{
		TemplateID:  "d-157cefc3327e4301b1a3ad065e35bb12",
		Categories:  []string{"Forgot Password"},
		DynamicData: make(map[string]interface{}),
	}, nil
}

func inviteMemberTemplate() (*providers.DynamicTemplate, error) {
	return &providers.DynamicTemplate{
		TemplateID:  "d-e5db06aece334ecdba891929dd1133cd",
		Categories:  []string{"Invite Member"},
		DynamicData: make(map[string]interface{}),
	}, nil
}

func contactUsTemplate() (*providers.DynamicTemplate, error) {
	return &providers.DynamicTemplate{
		TemplateID:  "d-162a7456fc274f08ba5b46ad1336b1e8",
		Categories:  []string{"Contact Us"},
		DynamicData: make(map[string]interface{}),
	}, nil
}

func subscribeTemplate() (*providers.DynamicTemplate, error) {
	return &providers.DynamicTemplate{
		TemplateID:  "d-0b42794559714f0d9cd7d56226b8255b",
		Categories:  []string{"Subscribe"},
		DynamicData: make(map[string]interface{}),
	}, nil
}
