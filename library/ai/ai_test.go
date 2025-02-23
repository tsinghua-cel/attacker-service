package aiattack

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func TestRex2(t *testing.T) {
	var content = "[\n  {\n    \"slot\": \"5\",\n    \"actions\": {\n      \"BlockGetNewParentRoot\": \"modifyParentRoot:3\",\n      \"BlockBeforeBroadCast\": \"delayWithDuration:2\",\n      \"AttestBeforeSign\": \"modifyAttestHead:3,modifyAttestSource:0,modifyAttestTarget:3\"\n    }\n  },\n  {\n    \"slot\": \"10\",\n    \"actions\": {\n      \"BlockBeforeSign\": \"packPooledAttest\",\n      \"AttestBeforeBroadCast\": \"delayWithDuration:3\",\n      \"AttestBeforeSign\": \"modifyAttestHead:5,modifyAttestTarget:5\"\n    }\n  },\n  {\n    \"slot\": \"15\",\n    \"actions\": {\n      \"BlockBeforeBroadCast\": \"return\",\n      \"AttestBeforeSign\": \"modifyAttestSource:5,modifyAttestTarget:15\"\n    }\n  },\n  {\n    \"slot\": \"35\",\n    \"actions\": {\n      \"BlockGetNewParentRoot\": \"modifyParentRoot:33\",\n      \"AttestBeforeBroadCast\": \"delayWithDuration:1\",\n      \"AttestBeforeSign\": \"modifyAttestHead:33,modifyAttestSource:33\"\n    }\n  },\n  {\n    \"slot\": \"40\",\n    \"actions\": {\n      \"BlockBeforeSign\": \"packPooledAttest\",\n      \"BlockBeforeBroadCast\": \"delayWithDuration:2\",\n      \"AttestBeforeSign\": \"modifyAttestTarget:35\"\n    }\n  },\n  {\n    \"slot\": \"65\",\n    \"actions\": {\n      \"BlockGetNewParentRoot\": \"modifyParentRoot:63\",\n      \"AttestBeforeBroadCast\": \"return\",\n      \"AttestBeforeSign\": \"modifyAttestHead:63,modifyAttestSource:63,modifyAttestTarget:65\"\n    }\n  },\n  {\n    \"slot\": \"70\",\n    \"actions\": {\n      \"BlockBeforeBroadCast\": \"delayWithDuration:3\",\n      \"AttestBeforeSign\": \"modifyAttestHead:65,modifyAttestTarget:70\"\n    }\n  },\n  {\n    \"slot\": \"75\",\n    \"actions\": {\n      \"BlockBeforeSign\": \"packPooledAttest\",\n      \"AttestBeforeBroadCast\": \"delayWithDuration:2\",\n      \"AttestBeforeSign\": \"modifyAttestSource:65,modifyAttestTarget:75\"\n    }\n  }\n]"
	content = strings.Replace(content, "\n", "", -1)
	re := regexp.MustCompile("[(.*?)]")
	jsonStr := re.FindStringSubmatch(content)
	if len(jsonStr) > 0 {
		fmt.Println("json=", jsonStr)
	} else {
		t.Fatalf("jsonStr is empty")
	}
}

func TestRex(t *testing.T) {
	content := "```json\n[\n  {\n    \"slot\": \"1\",\n    \"actions\": {\n      \"BlockBeforeBroadCast\": \"delayWithDuration:3\",\n      \"BlockGetNewParentRoot\": \"modifyParentRoot:0\",\n      \"AttestBeforeSign\": \"modifyAttestHead:0;modifyAttestSource:0;modifyAttestTarget:0\"\n    }\n  },\n  {\n    \"slot\": \"2\",\n    \"actions\": {\n      \"BlockBeforeBroadCast\": \"return\",\n      \"AttestBeforeBroadCast\": \"delayWithDuration:2\",\n      \"AttestBeforeSign\": \"modifyAttestSource:1;modifyAttestTarget:1\"\n    }\n  },\n  {\n    \"slot\": \"16\",\n    \"actions\": {\n      \"BlockGetNewParentRoot\": \"modifyParentRoot:15\",\n      \"BlockBeforeSign\": \"packPooledAttest\",\n      \"AttestBeforeSign\": \"modifyAttestHead:15;modifyAttestSource:15;modifyAttestTarget:15\"\n    }\n  },\n  {\n    \"slot\": \"17\",\n    \"actions\": {\n      \"BlockBeforeBroadCast\": \"delayWithDuration:1\",\n      \"AttestBeforeBroadCast\": \"delayWithDuration:3\",\n      \"AttestBeforeSign\": \"modifyAttestSource:16;modifyAttestTarget:16\"\n    }\n  },\n  {\n    \"slot\": \"32\",\n    \"actions\": {\n      \"BlockGetNewParentRoot\": \"modifyParentRoot:31\",\n      \"AttestBeforeSign\": \"modifyAttestHead:31;modifyAttestSource:31;modifyAttestTarget:31\"\n    }\n  },\n  {\n    \"slot\": \"33\",\n    \"actions\": {\n      \"BlockBeforeBroadCast\": \"return\",\n      \"AttestBeforeBroadCast\": \"delayWithDuration:2\",\n      \"AttestBeforeSign\": \"modifyAttestSource:32;modifyAttestTarget:32\"\n    }\n  },\n  {\n    \"slot\": \"48\",\n    \"actions\": {\n      \"BlockGetNewParentRoot\": \"modifyParentRoot:47\",\n      \"BlockBeforeSign\": \"packPooledAttest\",\n      \"AttestBeforeSign\": \"modifyAttestHead:47;modifyAttestSource:47;modifyAttestTarget:47\"\n    }\n  },\n  {\n    \"slot\": \"64\",\n    \"actions\": {\n      \"BlockBeforeBroadCast\": \"delayWithDuration:3\",\n      \"BlockGetNewParentRoot\": \"modifyParentRoot:63\",\n      \"AttestBeforeSign\": \"modifyAttestHead:63;modifyAttestSource:63;modifyAttestTarget:63\"\n    }\n  },\n  {\n    \"slot\": \"80\",\n    \"actions\": {\n      \"BlockBeforeBroadCast\": \"return\",\n      \"AttestBeforeBroadCast\": \"delayWithDuration:1\",\n      \"AttestBeforeSign\": \"modifyAttestSource:79;modifyAttestTarget:79\"\n    }\n  },\n  {\n    \"slot\": \"96\",\n    \"actions\": {\n      \"BlockGetNewParentRoot\": \"modifyParentRoot:95\",\n      \"BlockBeforeSign\": \"packPooledAttest\",\n      \"AttestBeforeSign\": \"modifyAttestHead:95;modifyAttestSource:95;modifyAttestTarget:95\"\n    }\n  }\n]\n```"
	content = strings.Replace(content, "```json", "", -1)
	content = strings.Replace(content, "```", "", -1)
	content = strings.Replace(content, "\n", "", -1)
	content = strings.TrimSpace(content)
	re := regexp.MustCompile("\\[.*\\]")
	jsonStr := re.FindString(content)
	if len(jsonStr) > 0 {
		fmt.Println("json=", jsonStr)
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
