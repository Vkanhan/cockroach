// Copyright 2022 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package admission

import (
	"fmt"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/testutils/datapathutils"
	"github.com/cockroachdb/cockroach/pkg/util/admission/admissionpb"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/datadriven"
	"github.com/cockroachdb/redact"
)

func TestDiskBandwidthLimiter(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	var dbl *diskBandwidthLimiter
	dblToString := func() string {
		return string(redact.Sprint(dbl))
	}

	datadriven.RunTest(t, datapathutils.TestDataPath(t, "disk_bandwidth_limiter"),
		func(t *testing.T, d *datadriven.TestData) string {
			switch d.Cmd {
			case "init":
				dbl = newDiskBandwidthLimiter()
				return dblToString()

			case "compute":
				var readBytes, writeBytes, intProvisionedBytes int
				d.ScanArgs(t, "int-read-bytes", &readBytes)
				d.ScanArgs(t, "int-write-bytes", &writeBytes)
				d.ScanArgs(t, "int-provisioned-bytes", &intProvisionedBytes)
				diskLoad := intervalDiskLoadInfo{
					intReadBytes:            int64(readBytes),
					intWriteBytes:           int64(writeBytes),
					intProvisionedDiskBytes: int64(intProvisionedBytes),
					elasticBandwidthMaxUtil: 0.9,
				}
				var regularTokensUsed, elasticTokensUsed int64
				d.ScanArgs(t, "regular-tokens-used", &regularTokensUsed)
				d.ScanArgs(t, "elastic-tokens-used", &elasticTokensUsed)
				usedTokens := [admissionpb.NumWorkClasses]diskTokens{
					{writeByteTokens: regularTokensUsed}, // regular
					{writeByteTokens: elasticTokensUsed}, // elastic
				}

				dbl.computeElasticTokens(diskLoad, usedTokens)
				return dblToString()

			default:
				return fmt.Sprintf("unknown command: %s", d.Cmd)
			}
		})
}
