package helper

import (
	"encoding/json"
	"testing"
)

func Test_ParseResourceIdErrors(t *testing.T) {
	resourceIds := []string{
		"",
		" /subscriptions/d7b0cf13-ddf7-43ea-81f1-6f659767a318 ",
		" /subscriptions/d7b0cf13-ddf7-43ea-81f1-6f659767a318/resourceGroups/foo?-rg",
		"/subscriptions/d7b0cf13-ddf7-43ea-81f1-6f659767a318",
		"/subscriptions/d7b0cf13-ddf7-43ea-81f1-6f659767a318/",
		"/subscriptions/d7b0cf13-ddf7-43ea-81f1-6f659767a318/providers/microsoft.authorization/roleDefinitions/4a9ae827-6dc8-4573-8ac7-8239d42aa03f",
		"/subscriptions/d7b0cf13-ddf7-43ea-81f1-6f659767a318/providers/microsoft.authorization/roleDefinitions/4a9ae827-6dc8-4573-8ac7-8239d42aa03f/",
		"/subscriptions/ad404ddd-36a5-4ea8-b3e3-681e77487a63/providers/Microsoft.Authorization/policyAssignments/myAssignment",
		"/subscriptions/ad404ddd-36a5-4ea8-b3e3-681e77487a63/providers/Microsoft.Authorization/policyAssignments/myAssignment/",
		"/subscriptions/d7b0cf13-ddf7-43ea-81f1-6f659767a318/resourceGroups/foo-rg/providers/Microsoft.Network/routeTables/testroute/xzy/foo/bar/foo",
		"/subscriptions/d7b0cf13-ddf7-43ea-81f1-6f659767a318/resourceGroups/foo-rg/providers/Microsoft.Network/routeTables/testroute/xzy",
		"/subscriptions/d7b0cf13-ddf7-43ea-81f1-6f659767a318/resourceGroups/foo-rg/providers/Microsoft.Network/routeTables/testroute/xzy/",
	}

	for _, resourceId := range resourceIds {
		assertInvalidResourceId(t, resourceId)
	}
}

func Test_ParseResourceId(t *testing.T) {
	resourceIds := map[string]string{
		"/subscriptions/d7b0cf13-ddf7-43ea-81f1-6f659767a318/resourceGroups/foo-rg":                                                   "/subscriptions/d7b0cf13-ddf7-43ea-81f1-6f659767a318/resourceGroups/foo-rg",
		"/subscriptions/d7b0cf13-ddf7-43ea-81f1-6f659767a318/resourceGroups/foo-rg/":                                                  "/subscriptions/d7b0cf13-ddf7-43ea-81f1-6f659767a318/resourceGroups/foo-rg",
		"/subscriptions/d7b0cf13-ddf7-43ea-81f1-6f659767a318/resourceGroups/FOO-rg/":                                                  "/subscriptions/d7b0cf13-ddf7-43ea-81f1-6f659767a318/resourceGroups/foo-rg",
		"/subscriptions/d7b0cf13-ddf7-43ea-81f1-6f659767a318/resourceGroups/foo-rg/providers/Microsoft.Network/routeTables/testroute": "/subscriptions/d7b0cf13-ddf7-43ea-81f1-6f659767a318/resourceGroups/foo-rg/providers/microsoft.network/routetables/testroute",
	}

	for resourceId, expected := range resourceIds {
		assertResourceId(t, expected, resourceId)
	}
}

func assertResourceId(t *testing.T, expect string, val string) {
	t.Helper()

	if info, err := ParseResourceID(val); err == nil {
		resourceId := info.ResourceID()
		if resourceId != expect {
			infoJson, _ := json.Marshal(info)
			t.Errorf("expected: \"%s\", got \"%s\" (%s)", expect, resourceId, infoJson)
		}
	} else {
		t.Errorf("unable to parse resourceid \"%s\": %s", val, err)
	}

}

func assertInvalidResourceId(t *testing.T, val string) {
	t.Helper()
	if info, err := ParseResourceID(val); err == nil {
		infoJson, _ := json.Marshal(info)
		t.Errorf("assumed invalid resourceid but was parsed successfull \"%s\": %s", val, infoJson)
	}
}
