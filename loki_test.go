package loki

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

var data = `{
    "status": "success",
    "data": {
        "resultType": "streams",
        "result": [
            {
                "stream": {
                    "filename": "/var/log/pods/default_esix-controller-7f88f5cd4c-bjnb2_7df9cee1-1eda-43e5-8eb3-03009bf46323/esix-controller/8.log",
                    "job": "default/esix",
                    "namespace": "default",
                    "app": "esix",
                    "component": "esix-controller",
                    "pod_template_hash": "7f88f5cd4c",
                    "release": "esix",
                    "stream": "stdout",
                    "container": "esix-controller",
                    "pod": "esix-controller-7f88f5cd4c-bjnb2"
                },
                "values": [
                    [
                        "1605238230041171278",
                        "\tat java.lang.Thread.run(Thread.java:834) ~[?:?]"
                    ],
                    [
                        "1605238230041166229",
                        "\tat akka.actor.LightArrayRevolverScheduler$$anon$3.run(LightArrayRevolverScheduler.scala:241) ~[40:com.typesafe.akka.actor:2.5.23]"
                    ],
                    [
                        "1605238230041159665",
                        "\tat akka.actor.LightArrayRevolverScheduler$$anon$3.nextTick(LightArrayRevolverScheduler.scala:289) ~[40:com.typesafe.akka.actor:2.5.23]"
                    ],
                    [
                        "1605238230041140438",
                        "\tat akka.actor.LightArrayRevolverScheduler$$anon$3.executeBucket$1(LightArrayRevolverScheduler.scala:285) ~[40:com.typesafe.akka.actor:2.5.23]"
                    ],
                    [
                        "1605238230041135478",
                        "\tat akka.actor.LightArrayRevolverScheduler$TaskHolder.executeTask(LightArrayRevolverScheduler.scala:334) ~[40:com.typesafe.akka.actor:2.5.23]"
                    ],
                    [
                        "1605238230041130619",
                        "\tat scala.concurrent.Future$InternalCallbackExecutor$.execute(Future.scala:872) ~[392:org.scala-lang.scala-library:2.12.8.v20181128-140630-VFINAL-38cd84d]"
                    ],
                    [
                        "1605238230041125626",
                        "\tat scala.concurrent.BatchingExecutor.execute$(BatchingExecutor.scala:107) ~[392:org.scala-lang.scala-library:2.12.8.v20181128-140630-VFINAL-38cd84d]"
                    ],
                    [
                        "1605238230041120605",
                        "\tat scala.concurrent.BatchingExecutor.execute(BatchingExecutor.scala:113) ~[392:org.scala-lang.scala-library:2.12.8.v20181128-140630-VFINAL-38cd84d]"
                    ],
                    [
                        "1605238230041115686",
                        "\tat scala.concurrent.Future$InternalCallbackExecutor$.unbatchedExecute(Future.scala:874) ~[392:org.scala-lang.scala-library:2.12.8.v20181128-140630-VFINAL-38cd84d]"
                    ],
                    [
                        "1605238230041110967",
                        "\tat akka.actor.Scheduler$$anon$4.run(Scheduler.scala:202) ~[40:com.typesafe.akka.actor:2.5.23]"
                    ]
                ]
            }
        ],
        "stats": {
            "summary": {
                "bytesProcessedPerSecond": 124756024,
                "linesProcessedPerSecond": 723450,
                "totalBytesProcessed": 642188,
                "totalLinesProcessed": 3724,
                "execTime": 0.005147551
            },
            "store": {
                "totalChunksRef": 1,
                "totalChunksDownloaded": 1,
                "chunksDownloadTime": 0.000506846,
                "headChunkBytes": 0,
                "headChunkLines": 0,
                "decompressedBytes": 457792,
                "decompressedLines": 2562,
                "compressedBytes": 23654,
                "totalDuplicates": 0
            },
            "ingester": {
                "totalReached": 1,
                "totalChunksMatched": 1,
                "totalBatches": 0,
                "totalLinesSent": 0,
                "headChunkBytes": 184396,
                "headChunkLines": 1162,
                "decompressedBytes": 0,
                "decompressedLines": 0,
                "compressedBytes": 0,
                "totalDuplicates": 0
            }
        }
    }
}`

func TestLokiStruct(t *testing.T) {
	var loki Loki
	err := json.Unmarshal([]byte(data), &loki)
	if err != nil {
		t.Fatal(err)
	}
	for _, data := range loki.Data.Result {
		fmt.Println(data)
	}
}

func TestLokiClient_LeastLogs(t *testing.T) {
	lc := NewLokiClient("10.1.14.136", 3100)
	l, err := lc.LogsRange(1000, BACKWARD, &Range{
		Start:  time.Date(2020, 10, 1, 1, 1, 1, 1, time.Local),
		End:    time.Now(),
		Enable: true,
	}, "{namespace=\"default\"}|~\"error\"")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(l)
}

func BenchmarkLokiClient_LeastLogs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lc := NewLokiClient("10.1.14.136", 3100)
		_, err := lc.LogsLeast(10, FORWARD, "{namespace=\"default\"}|~\"log_type\"|~\"system\"")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestBaseURL(t *testing.T) {
	var u = NewURLBuilder()
	b , err := u.BaseUrl("http","10.1.14.136", 3100)
	if err != nil{
		t.Fatal(err)
	}
	b.Query()
	t.Log(u)
}
