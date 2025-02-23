package aiattack

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func TestRex(t *testing.T) {
	content := "```json\n[\n  {\n    \"slot\": \"1\",\n    \"actions\": {\n      \"BlockBeforeBroadCast\": \"delayWithDuration:3\",\n      \"BlockGetNewParentRoot\": \"modifyParentRoot:0\",\n      \"AttestBeforeSign\": \"modifyAttestHead:0;modifyAttestSource:0;modifyAttestTarget:0\"\n    }\n  },\n  {\n    \"slot\": \"2\",\n    \"actions\": {\n      \"BlockBeforeBroadCast\": \"return\",\n      \"AttestBeforeBroadCast\": \"delayWithDuration:2\",\n      \"AttestBeforeSign\": \"modifyAttestSource:1;modifyAttestTarget:1\"\n    }\n  },\n  {\n    \"slot\": \"16\",\n    \"actions\": {\n      \"BlockGetNewParentRoot\": \"modifyParentRoot:15\",\n      \"BlockBeforeSign\": \"packPooledAttest\",\n      \"AttestBeforeSign\": \"modifyAttestHead:15;modifyAttestSource:15;modifyAttestTarget:15\"\n    }\n  },\n  {\n    \"slot\": \"17\",\n    \"actions\": {\n      \"BlockBeforeBroadCast\": \"delayWithDuration:1\",\n      \"AttestBeforeBroadCast\": \"delayWithDuration:3\",\n      \"AttestBeforeSign\": \"modifyAttestSource:16;modifyAttestTarget:16\"\n    }\n  },\n  {\n    \"slot\": \"32\",\n    \"actions\": {\n      \"BlockGetNewParentRoot\": \"modifyParentRoot:31\",\n      \"AttestBeforeSign\": \"modifyAttestHead:31;modifyAttestSource:31;modifyAttestTarget:31\"\n    }\n  },\n  {\n    \"slot\": \"33\",\n    \"actions\": {\n      \"BlockBeforeBroadCast\": \"return\",\n      \"AttestBeforeBroadCast\": \"delayWithDuration:2\",\n      \"AttestBeforeSign\": \"modifyAttestSource:32;modifyAttestTarget:32\"\n    }\n  },\n  {\n    \"slot\": \"48\",\n    \"actions\": {\n      \"BlockGetNewParentRoot\": \"modifyParentRoot:47\",\n      \"BlockBeforeSign\": \"packPooledAttest\",\n      \"AttestBeforeSign\": \"modifyAttestHead:47;modifyAttestSource:47;modifyAttestTarget:47\"\n    }\n  },\n  {\n    \"slot\": \"64\",\n    \"actions\": {\n      \"BlockBeforeBroadCast\": \"delayWithDuration:3\",\n      \"BlockGetNewParentRoot\": \"modifyParentRoot:63\",\n      \"AttestBeforeSign\": \"modifyAttestHead:63;modifyAttestSource:63;modifyAttestTarget:63\"\n    }\n  },\n  {\n    \"slot\": \"80\",\n    \"actions\": {\n      \"BlockBeforeBroadCast\": \"return\",\n      \"AttestBeforeBroadCast\": \"delayWithDuration:1\",\n      \"AttestBeforeSign\": \"modifyAttestSource:79;modifyAttestTarget:79\"\n    }\n  },\n  {\n    \"slot\": \"96\",\n    \"actions\": {\n      \"BlockGetNewParentRoot\": \"modifyParentRoot:95\",\n      \"BlockBeforeSign\": \"packPooledAttest\",\n      \"AttestBeforeSign\": \"modifyAttestHead:95;modifyAttestSource:95;modifyAttestTarget:95\"\n    }\n  }\n]\n```"
	content = strings.Replace(content, "\n", "", -1)
	re := regexp.MustCompile("```json(.*?)```")
	jsonStr := re.FindStringSubmatch(content)
	if len(jsonStr) > 1 {
		fmt.Println("json=", jsonStr[1])
	} else {
		t.Fatalf("jsonStr is empty")
	}
	//re := regexp.MustCompile("```json(.*?)```")
	////re := regexp.MustCompile("```json\n(.*?)```")
	//jsonStr := re.FindString(content)
	//fmt.Println("json=", jsonStr)
	//if len(jsonStr) == 0 {
	//	t.Fatalf("jsonStr is empty")
	//}
}
