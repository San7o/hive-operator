/*
                    GNU GENERAL PUBLIC LICENSE
                       Version 2, June 1991

 Copyright (C) 1989, 1991 Free Software Foundation, Inc.,
 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.
*/

// SPDX-License-Identifier: GPL-2.0-only

package controller

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/client"

	hivebpf "github.com/San7o/hive-operator/internal/controller/ebpf"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

func Output(client client.Reader) {

	ctx := context.Background()
	log := logger.FromContext(ctx)

	for {
		alert, err := hivebpf.ReadAlert(ctx, client)
		if err != nil {
			log.Error(err, "Output Error Read alert")
			continue
		}

		jsonAlert, err := json.Marshal(alert)
		if err != nil {
			log.Error(err, "Output Error Json Marshal")
			continue
		}
		if alert.Metadata.Callback != "" {
			_, err := http.Post(alert.Metadata.Callback, "application/json", bufio.NewReader(bytes.NewReader(jsonAlert)))
			if err != nil {
				log.Error(err, "Output Error Post Callback")
			}
		} else {
			log.Info("Access Detected", "HiveAlert", string(jsonAlert))
		}
	}
}
