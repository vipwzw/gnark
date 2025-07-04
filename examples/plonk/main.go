// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package main

import (
	"fmt"
	"log"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/plonk"
	cs "github.com/consensys/gnark/constraint/bn254"
	"github.com/consensys/gnark/frontend/cs/scs"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/test/unsafekzg"
)

// In this example we show how to use PLONK with KZG commitments. The circuit that is
// showed here is the same as in ../exponentiate.

// Circuit y == x**e
// only the bitSize least significant bits of e are used
type Circuit struct {
	// tagging a variable is optional
	// default uses variable name and secret visibility.
	X frontend.Variable `gnark:",public"`
	Y frontend.Variable `gnark:",public"`

	E frontend.Variable
}

// Define declares the circuit's constraints
// y == x**e
func (circuit *Circuit) Define(api frontend.API) error {
	// number of bits of exponent
	const bitSize = 4000
	// specify constraints
	output := frontend.Variable(1)
	bits := api.ToBinary(circuit.E, bitSize)
	base := circuit.X
	for i := 0; i < len(bits); i++ {
		output = api.Select(bits[i], api.Mul(output, base), output)
		base = api.Mul(base, base)
	}
	api.AssertIsEqual(circuit.Y, output)
	return nil
}

func (circuit *Circuit) Define2(api frontend.API) error {
	const bitSize = 32
	output := frontend.Variable(1)
	bits := api.ToBinary(circuit.E, bitSize)

	for i := 0; i < len(bits); i++ {
		api.Println(fmt.Sprintf("e[%d]", i), bits[i]) // we may print a variable for testing and / or debugging purposes
		if i != 0 {
			output = api.Mul(output, output)
		}
		api.Println("out1->", output)
		multiply := api.Mul(output, circuit.X)
		output = api.Select(bits[len(bits)-1-i], multiply, output)
		api.Println("out2->", output, "multiply->", multiply, "X->", circuit.X, "Y->", circuit.Y, "E->", circuit.E)
	}
	api.Println("Y = ", circuit.Y, "output = ", output)
	api.AssertIsEqual(circuit.Y, output)
	return nil
}

func intToBinaryBits(n int) (bits []byte) {
	for n > 0 {
		// 将n的最低位添加到bits字符串
		bits = append(bits, byte(n&1))
		// 右移n，以检查下一个位
		n >>= 1
	}
	return bits
}

func my_test_pow(x, e int) int {
	output := 1
	bits := intToBinaryBits(e)
	base := x
	for i := 0; i < len(bits); i++ {
		if bits[i] == 1 {
			output = output * base
		}
		base *= base
	}
	return output
}

func test_pow(x, e int) int {
	/*
		// specify constraints
		output := frontend.Variable(1)
		bits := api.ToBinary(circuit.E, bitSize)

		for i := 0; i < len(bits); i++ {
			// api.Println(fmt.Sprintf("e[%d]", i), bits[i]) // we may print a variable for testing and / or debugging purposes

			if i != 0 {
				output = api.Mul(output, output)
			}
			multiply := api.Mul(output, circuit.X)
			output = api.Select(bits[len(bits)-1-i], multiply, output)

		}

		api.AssertIsEqual(circuit.Y, output)
	*/
	output := 1
	bits := intToBinaryBits(e)
	fmt.Println("bits:", bits)
	for i := 0; i < len(bits); i++ {
		if i != 0 {
			output = output * output
		}
		multiply := output * x
		if bits[i] == 1 {
			output = multiply
		}
	}
	return output
}

func main() {
	/*
		r := test_pow(2, 10)
		fmt.Println("2^ 10 = ", r)
		r = my_test_pow(2, 10)
		fmt.Println("2^ 10 = ", r)
		return
	*/
	var circuit Circuit
	// // building the circuit...
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, &circuit)
	if err != nil {
		fmt.Println("circuit compilation error")
	}

	// create the necessary data for KZG.
	// This is a toy example, normally the trusted setup to build ZKG
	// has been run before.
	// The size of the data in KZG should be the closest power of 2 bounding //
	// above max(nbConstraints, nbVariables).
	scs := ccs.(*cs.SparseR1CS)
	srs, srsLagrange, err := unsafekzg.NewSRS(scs)
	if err != nil {
		panic(err)
	}

	// Correct data: the proof passes
	{
		// Witnesses instantiation. Witness is known only by the prover,
		// while public w is a public data known by the verifier.
		var w Circuit
		w.X = 2
		w.E = 10
		w.Y = 1024

		witnessFull, err := frontend.NewWitness(&w, ecc.BN254.ScalarField())
		if err != nil {
			log.Fatal(err)
		}

		witnessPublic, err := frontend.NewWitness(&w, ecc.BN254.ScalarField(), frontend.PublicOnly())
		if err != nil {
			log.Fatal(err)
		}

		// public data consists of the polynomials describing the constants involved
		// in the constraints, the polynomial describing the permutation ("grand
		// product argument"), and the FFT domains.
		pk, vk, err := plonk.Setup(ccs, srs, srsLagrange)
		//_, err := plonk.Setup(r1cs, kate, &publicWitness)
		if err != nil {
			log.Fatal(err)
		}

		proof, err := plonk.Prove(ccs, pk, witnessFull)
		if err != nil {
			log.Fatal(err)
		}

		err = plonk.Verify(proof, vk, witnessPublic)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Wrong data: the proof fails
	{
		// Witnesses instantiation. Witness is known only by the prover,
		// while public w is a public data known by the verifier.
		var w, pW Circuit
		w.X = 2
		w.E = 12
		w.Y = 4096

		pW.X = 3
		pW.Y = 4096

		witnessFull, err := frontend.NewWitness(&w, ecc.BN254.ScalarField())
		if err != nil {
			log.Fatal(err)
		}

		witnessPublic, err := frontend.NewWitness(&pW, ecc.BN254.ScalarField(), frontend.PublicOnly())
		if err != nil {
			log.Fatal(err)
		}

		// public data consists of the polynomials describing the constants involved
		// in the constraints, the polynomial describing the permutation ("grand
		// product argument"), and the FFT domains.
		pk, vk, err := plonk.Setup(ccs, srs, srsLagrange)
		//_, err := plonk.Setup(r1cs, kate, &publicWitness)
		if err != nil {
			log.Fatal(err)
		}

		proof, err := plonk.Prove(ccs, pk, witnessFull)
		if err != nil {
			log.Fatal(err)
		}

		err = plonk.Verify(proof, vk, witnessPublic)
		if err == nil {
			log.Fatal("Error: wrong proof is accepted")
		}
	}

}
