// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by gnark DO NOT EDIT

package plonk_test

import (
	"testing"

	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/internal/backend/circuits"
	curve "github.com/consensys/gurvy/bw761"
)

func TestCircuits(t *testing.T) {
	for name, circuit := range circuits.Circuits {

		t.Run(name, func(t *testing.T) {
			assert := plonk.NewAssert(t)
			pcs, err := frontend.Compile(curve.ID, backend.PLONK, circuit.Circuit)
			assert.NoError(err)
			assert.ProverSucceeded(pcs, circuit.Good)
			assert.ProverFailed(pcs, circuit.Bad)
		})
	}
}
