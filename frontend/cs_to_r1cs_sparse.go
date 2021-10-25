/*
Copyright © 2020 ConsenSys

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package frontend

import (
	"math/big"
	"sort"
	"sync"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/internal/backend/compiled"

	bls12377r1cs "github.com/consensys/gnark/internal/backend/bls12-377/cs"
	bls12381r1cs "github.com/consensys/gnark/internal/backend/bls12-381/cs"
	bls24315r1cs "github.com/consensys/gnark/internal/backend/bls24-315/cs"
	bn254r1cs "github.com/consensys/gnark/internal/backend/bn254/cs"
	bw6761r1cs "github.com/consensys/gnark/internal/backend/bw6-761/cs"
)

// sparseR1CS extends the ConstraintSystem
// alongside with some intermediate data structures needed to convert from
// ConstraintSystem representataion to SparseR1CS
type sparseR1CS struct {
	*constraintSystem

	ccs compiled.SparseR1CS

	// we start our internal variables counting after the ConstraintSystem index
	// when we process R1C linear expressions
	// this will create new internal wires in the SparseR1CS
	// we add these new wires starting at position len(previousInternalWires)
	scsInternalVariables int

	// keep track of solved variables to split the R1C in a sensible manner
	// and guarantee that the solver will encounter at most one unsolved wire
	// per SparseR1C
	solvedVariables []bool

	currentR1CDebugID int // mark the current R1C debugID

	// map LinearExpression -> Term. The goal is to not reduce
	// the same linear expression twice.
	record map[uint64][]innerRecord
}

type innerRecord struct {
	t compiled.Term
	l compiled.LinearExpression
}

var bOne = new(big.Int).SetInt64(1)

func (cs *constraintSystem) toSparseR1CS(curveID ecc.ID) (CompiledConstraintSystem, error) {

	res := sparseR1CS{
		constraintSystem: cs,
		ccs: compiled.SparseR1CS{
			CS: compiled.CS{
				NbInternalVariables: len(cs.internal.variables),
				NbPublicVariables:   len(cs.public.variables) - 1, // the ONE_WIRE is discarded in PlonK
				NbSecretVariables:   len(cs.secret.variables),
				DebugInfo:           make([]compiled.LogEntry, len(cs.debugInfo)),
				Logs:                make([]compiled.LogEntry, len(cs.logs)),
				MDebug:              make(map[int]int),
				MHints:              make(map[int]compiled.Hint),
			},
			Constraints: make([]compiled.SparseR1C, 0, len(cs.constraints)),
		},
		solvedVariables:      make([]bool, len(cs.internal.variables), len(cs.internal.variables)*2),
		scsInternalVariables: len(cs.internal.variables),
		currentR1CDebugID:    -1,
		record:               make(map[uint64][]innerRecord, len(cs.internal.variables)),
	}

	// logs, debugInfo and hints are copied, the only thing that will change
	// is that ID of the wires will be offseted to take into account the final wire vector ordering
	// that is: public wires  | secret wires | internal wires

	// we mark hint wires are solved
	// each R1C from the frontend.ConstraintSystem is allowed to have at most one unsolved wire
	// excluding hints. We mark hint wires as "solved" to ensure spliting R1C to SparseR1C
	// won't create invalid SparseR1C constraint with more than one wire to solve for the solver
	for vID := range cs.mHints {
		res.solvedVariables[vID] = true
	}

	// convert the R1C to SparseR1C
	// in particular, all linear expressions that appear in the R1C
	// will be split in multiple constraints in the SparseR1C
	for i := 0; i < len(cs.constraints); i++ {
		// we set currentR1CDebugID to the debugInfo ID corresponding to the R1C we're processing
		// if present. All constraints created throuh addConstraint will add a new mapping
		if dID, ok := cs.mDebug[i]; ok {
			res.currentR1CDebugID = dID
		} else {
			res.currentR1CDebugID = -1
		}
		res.r1cToSparseR1C(cs.constraints[i])
	}

	// shift variable ID
	// we want publicWires | privateWires | internalWires
	shiftVID := func(oldID int, visibility compiled.Visibility) int {
		switch visibility {
		case compiled.Internal:
			return oldID + res.ccs.NbPublicVariables + res.ccs.NbSecretVariables
		case compiled.Public:
			return oldID - 1
		case compiled.Secret:
			return oldID + res.ccs.NbPublicVariables
		default:
			return oldID
		}
	}

	offsetTermID := func(t *compiled.Term) {
		if *t == 0 {
			// in a PLONK constraint, not all terms are necessarily set,
			// the terms which are not set are equal to zero. We just
			// need to skip them.
			return
		}
		_, vID, visibility := t.Unpack()
		if vID == 0 && visibility == compiled.Public {
			// this would not happen in a plonk constraint as the constant term has been popped out
			// however it may happen in the logs or the hints that contains
			// terms associated with the ONE wire
			// workaround; we set the visibility to Virtual so that the solver recognizes that as a constant
			t.SetVariableVisibility(compiled.Virtual)
			return
		}
		t.SetVariableID(shiftVID(vID, visibility))
	}

	// offset the IDs of all constraints so that the variables are
	// numbered like this: [publicVariables | secretVariables | internalVariables ]
	for i := 0; i < len(res.ccs.Constraints); i++ {
		r1c := &res.ccs.Constraints[i]
		// offset each term in the constraint
		offsetTermID(&r1c.L)
		offsetTermID(&r1c.R)
		offsetTermID(&r1c.O)
		offsetTermID(&r1c.M[0])
		offsetTermID(&r1c.M[1])
	}

	// we need to offset the ids in logs & debugInfo
	for i := 0; i < len(cs.logs); i++ {
		res.ccs.Logs[i] = compiled.LogEntry{
			Format:    cs.logs[i].Format,
			ToResolve: make([]compiled.Term, len(cs.logs[i].ToResolve)),
		}
		copy(res.ccs.Logs[i].ToResolve, cs.logs[i].ToResolve)

		for j := 0; j < len(res.ccs.Logs[i].ToResolve); j++ {
			offsetTermID(&res.ccs.Logs[i].ToResolve[j])
		}
	}
	for i := 0; i < len(cs.debugInfo); i++ {
		res.ccs.DebugInfo[i] = compiled.LogEntry{
			Format:    cs.debugInfo[i].Format,
			ToResolve: make([]compiled.Term, len(cs.debugInfo[i].ToResolve)),
		}
		copy(res.ccs.DebugInfo[i].ToResolve, cs.debugInfo[i].ToResolve)

		for j := 0; j < len(res.ccs.DebugInfo[i].ToResolve); j++ {
			offsetTermID(&res.ccs.DebugInfo[i].ToResolve[j])
		}
	}

	// we need to offset the ids in the hints
	for vID, hint := range cs.mHints {
		k := shiftVID(vID, compiled.Internal)
		inputs := make([]compiled.LinearExpression, len(hint.Inputs))
		copy(inputs, hint.Inputs)
		for j := 0; j < len(inputs); j++ {
			for k := 0; k < len(inputs[j]); k++ {
				offsetTermID(&inputs[j][k])
			}
		}
		res.ccs.MHints[k] = compiled.Hint{ID: hint.ID, Inputs: inputs}
	}

	// update number of internal variables with new wires created
	// while processing R1C -> SparseR1C
	res.ccs.NbInternalVariables = res.scsInternalVariables

	switch curveID {
	case ecc.BLS12_377:
		return bls12377r1cs.NewSparseR1CS(res.ccs, cs.coeffs), nil
	case ecc.BLS12_381:
		return bls12381r1cs.NewSparseR1CS(res.ccs, cs.coeffs), nil
	case ecc.BN254:
		return bn254r1cs.NewSparseR1CS(res.ccs, cs.coeffs), nil
	case ecc.BW6_761:
		return bw6761r1cs.NewSparseR1CS(res.ccs, cs.coeffs), nil
	case ecc.BLS24_315:
		return bls24315r1cs.NewSparseR1CS(res.ccs, cs.coeffs), nil
	default:
		panic("unknown curveID")
	}

}

// findUnsolvedVariable returns the variable to solve in the r1c. The variables
// which are not internal are considered solved, otherwise the solvedVariables
// slice hold the record of which variables have been solved.
func findUnsolvedVariable(r1c compiled.R1C, solvedVariables []bool) (int, int) {
	// find the variable to solve among L,R,O. pos=0,1,2 corresponds to left,right,o.
	for i := 0; i < len(r1c.L); i++ {
		_, vID, visibility := r1c.L[i].Unpack()
		if visibility == compiled.Internal && !solvedVariables[vID] {
			return 0, vID
		}
	}
	for i := 0; i < len(r1c.R); i++ {
		_, vID, visibility := r1c.R[i].Unpack()
		if visibility == compiled.Internal && !solvedVariables[vID] {
			return 1, vID
		}
	}
	for i := 0; i < len(r1c.O); i++ {
		_, vID, visibility := r1c.O[i].Unpack()
		if visibility == compiled.Internal && !solvedVariables[vID] {
			return 2, vID
		}
	}
	return -1, -1
}

// returns l with the term (id+coef) holding the id-th variable removed
// No side effects on l.
func popInternalVariable(l compiled.LinearExpression, id int) (compiled.LinearExpression, compiled.Term) {
	var t compiled.Term
	_l := make([]compiled.Term, len(l)-1)
	c := 0
	for i := 0; i < len(l); i++ {
		v := l[i]
		if v.VariableVisibility() == compiled.Internal && v.VariableID() == id {
			t = v
			continue
		}
		_l[c] = v
		c++
	}
	return _l, t
}

// returns ( b/computeGCD(b...), computeGCD(b...) )
// if gcd is != 0 and gcd != 1, returns true
func (scs *sparseR1CS) computeGCD(l compiled.LinearExpression, gcd *big.Int) {
	gcd.SetUint64(0)
	for i := 0; i < len(l); i++ {
		cID := l[i].CoeffID()
		if cID == compiled.CoeffIdZero {
			continue
		}
		gcd.GCD(nil, nil, gcd, &scs.coeffs[cID])

		if gcd.IsUint64() && gcd.Uint64() == 1 {
			break
		}
	}

}

// return true if linear expression contains one or minusOne coefficients
func hasOnes(l compiled.LinearExpression) bool {
	for i := 0; i < len(l); i++ {
		cID := l[i].CoeffID()
		if cID == compiled.CoeffIdMinusOne || cID == compiled.CoeffIdOne {
			return true
		}
	}
	return false
}

// reduce sets gcd = gcd(l.coefs) and returns l/gcd(l.coefs)
// if gcd == 1, this returns l
func (scs *sparseR1CS) reduce(l compiled.LinearExpression, gcd *big.Int) compiled.LinearExpression {
	mustNeg := scs.coeffs[l[0].CoeffID()].Sign() == -1

	// fast path: if any of the coeffs is 1 or -1, no need to compute the GCD
	if hasOnes(l) {
		if !mustNeg {
			gcd.SetUint64(1)
			return l
		}
		gcd.SetInt64(-1)
		return scs.divideLinearExpression(l, gcd)
	}

	// compute gcd
	scs.computeGCD(l, gcd)

	if mustNeg {
		// ensure the gcd doesn't depend on the sign
		gcd.Neg(gcd)
	}

	if gcd.IsUint64() && (gcd.Uint64() == 0 || gcd.Uint64() == 1) {
		// no need to create a new linear expression
		return l
	}

	return scs.divideLinearExpression(l, gcd)

}

func (scs *sparseR1CS) divideLinearExpression(l compiled.LinearExpression, d *big.Int) compiled.LinearExpression {
	// copy linear expression
	r := make(compiled.LinearExpression, len(l))
	copy(r, l)

	// new coeff
	lambda := bigIntPool.Get().(*big.Int)

	if d.IsInt64() && d.Int64() == -1 {
		for i := 0; i < len(r); i++ {
			cID := r[i].CoeffID()
			if cID == compiled.CoeffIdZero {
				continue
			}
			lambda.Neg(&scs.coeffs[cID])
			r[i].SetCoeffID(scs.coeffID(lambda))
		}
	} else {
		for i := 0; i < len(r); i++ {
			cID := r[i].CoeffID()
			if cID == compiled.CoeffIdZero {
				continue
			}
			lambda.Div(&scs.coeffs[cID], d)
			r[i].SetCoeffID(scs.coeffID(lambda))
		}
	}

	bigIntPool.Put(lambda)
	return r
}

// pops the constant associated to the one_wire in the cs, which will become
// a constant in a PLONK constraint.
//
// Returns the reduced linear expression and the ID of the coeff corresponding to the constant term (in cs.coeffs).
// If there is no constant term, the id is 0 (the 0-th entry is reserved for this purpose).
//
// ex: if l = <expr> + k1*ONE_WIRE the function returns <expr>, k1.
func (scs *sparseR1CS) popConstantTerm(l compiled.LinearExpression) (compiled.LinearExpression, big.Int) {

	const idOneWire = 0

	for i := 0; i < len(l); i++ {
		if l[i].VariableID() == idOneWire && l[i].VariableVisibility() == compiled.Public {
			lCopy := make(compiled.LinearExpression, len(l)-1)
			copy(lCopy, l[:i])
			copy(lCopy[i:], l[i+1:])
			return lCopy, scs.coeffs[l[i].CoeffID()]
		}
	}

	return l, big.Int{}
}

// newTerm creates a new term =coeff*new_variable and records it in the scs
// if idCS is set, uses it as variable id and does not increment the number
// of new internal variables created
func (scs *sparseR1CS) newTerm(coeff *big.Int, idCS ...int) compiled.Term {
	var vID int
	if len(idCS) > 0 {
		vID = idCS[0]
		scs.solvedVariables[vID] = true
	} else {
		vID = scs.scsInternalVariables
		scs.scsInternalVariables++
		scs.solvedVariables = append(scs.solvedVariables, true)
	}

	return compiled.Pack(vID, scs.coeffID(coeff), compiled.Internal)
}

// addConstraint records a plonk constraint in the ccs
// The function ensures that all variables ID are set, even
// if the corresponding coefficients are 0.
// A plonk constraint will always look like this:
// L+R+L.R+O+K = 0
func (scs *sparseR1CS) addConstraint(c compiled.SparseR1C) {
	// ensure wire(L) == wire(M[0]) && wire(R) == wire(M[1])
	if c.L == 0 {
		c.L.SetVariableID(c.M[0].VariableID())
	}
	if c.R == 0 {
		c.R.SetVariableID(c.M[1].VariableID())
	}
	if c.M[0] == 0 {
		c.M[0].SetVariableID(c.L.VariableID())
	}
	if c.M[1] == 0 {
		c.M[1].SetVariableID(c.R.VariableID())
	}
	if scs.currentR1CDebugID != -1 {
		scs.ccs.MDebug[len(scs.ccs.Constraints)] = scs.currentR1CDebugID
	}
	scs.ccs.Constraints = append(scs.ccs.Constraints, c)
}

// if t=a*variable, it returns -a*variable
func (scs *sparseR1CS) negate(t compiled.Term) compiled.Term {
	// non existing term are zero, if we negate it it's no
	// longer zero and checks to see if a variable exist will
	// fail (ex: in r1cToPlonkConstraint we might call negate
	// on non existing variables, when split is called with
	// le = nil)
	if t == 0 {
		return t
	}
	cID := t.CoeffID()
	switch cID {
	case compiled.CoeffIdMinusOne:
		t.SetCoeffID(compiled.CoeffIdOne)
	case compiled.CoeffIdZero:
		// do nothing.
	case compiled.CoeffIdOne:
		t.SetCoeffID(compiled.CoeffIdMinusOne)
	default:
		coeff := bigIntPool.Get().(*big.Int)
		coeff.Neg(&scs.coeffs[t.CoeffID()])
		t.SetCoeffID(scs.coeffID(coeff))
		bigIntPool.Put(coeff)
	}

	return t
}

// multiplies t by the provided coefficient
func (scs *sparseR1CS) multiply(t compiled.Term, c *big.Int) compiled.Term {
	// fast path
	if c.IsInt64() {
		v := c.Int64()
		switch v {
		case 0:
			t.SetCoeffID(compiled.CoeffIdZero)
			return t
		case 1:
			return t
		case -1:

			switch t.CoeffID() {
			case compiled.CoeffIdZero:
				return t
			case compiled.CoeffIdOne:
				t.SetCoeffID(compiled.CoeffIdMinusOne)
				return t
			case compiled.CoeffIdMinusOne:
				t.SetCoeffID(compiled.CoeffIdOne)
				return t
			}
		}
	}
	coeff := bigIntPool.Get().(*big.Int)
	coeff.Mul(&scs.coeffs[t.CoeffID()], c)
	t.SetCoeffID(scs.coeffID(coeff))
	bigIntPool.Put(coeff)
	return t
}

func (scs *sparseR1CS) getRecord(l compiled.LinearExpression) (compiled.Term, bool) {
	list, ok := scs.record[l.Hash()]
	if !ok {
		return 0, false
	}

	for i := 0; i < len(list); i++ {
		if list[i].l.Equal(l) {
			return list[i].t, true
		}
	}

	return 0, false
}

func (scs *sparseR1CS) putRecord(l compiled.LinearExpression, t compiled.Term) {
	id := l.Hash()
	list := scs.record[id]
	list = append(list, innerRecord{t: t, l: l})
	scs.record[id] = list
}

func (scs *sparseR1CS) split(l compiled.LinearExpression) compiled.Term {

	// floor case
	if len(l) == 1 {
		return l[0]
	}

	lGCD := bigIntPool.Get().(*big.Int)
	// check if l is recorded, if so we get it from the record
	lReduced := scs.reduce(l, lGCD)
	if t, ok := scs.getRecord(lReduced); ok {
		t.SetCoeffID(scs.coeffID(lGCD))
		bigIntPool.Put(lGCD)
		return t
	}

	// find if in the left side the constraint is recorded
	gcd := bigIntPool.Get().(*big.Int)

	for i := len(l) - 1; i > 0; i-- {
		ll := scs.reduce(lReduced[:i], gcd)
		if t, ok := scs.getRecord(ll); ok {
			t = scs.multiply(t, gcd)
			o := scs.newTerm(bOne)
			b := scs.split(lReduced[i:])
			scs.addConstraint(compiled.SparseR1C{L: t, R: b, O: scs.negate(o)})
			scs.putRecord(lReduced, o)
			r := scs.multiply(o, lGCD)
			bigIntPool.Put(lGCD)
			bigIntPool.Put(gcd)
			return r
		}
	}
	bigIntPool.Put(gcd)

	// else we build the reduction starting from l[0]
	o := scs.newTerm(bOne)
	a := lReduced[0]
	b := scs.split(lReduced[1:])
	scs.addConstraint(compiled.SparseR1C{L: a, R: b, O: scs.negate(o)})
	scs.putRecord(lReduced, o)
	r := scs.multiply(o, lGCD)
	bigIntPool.Put(lGCD)

	return r
}

// r1cToSparseR1C splits a r1c constraint
func (scs *sparseR1CS) r1cToSparseR1C(r1c compiled.R1C) {

	// find if the variable to solve is in the left, right, or o linear expression
	lro, idCS := findUnsolvedVariable(r1c, scs.solvedVariables)
	if lro == -1 {
		// this may happen if a constraint contained hint wires, that are marked as solved.
		// or if we r1c is an assertion (ie it does not yield any output)
		scs.splitR1C(r1c)
		return // no variable to solve here.
	}

	l := r1c.L
	r := r1c.R
	o := r1c.O
	sort.Sort(l)
	sort.Sort(r)
	sort.Sort(o)

	// if the unsolved variable in not in o,
	// ensure that it is in r1c.L
	if lro == 1 {
		l, r = r, l
		lro = 0
	}

	var (
		cK big.Int // constant K
		cS big.Int // constant S (associated with toSolve)
	)
	var toSolve compiled.Term

	l, cL := scs.popConstantTerm(l)
	r, cR := scs.popConstantTerm(r)
	o, cO := scs.popConstantTerm(o)

	// pop the unsolved wire from the linearexpression
	if lro == 0 { // unsolved is in L
		l, toSolve = popInternalVariable(l, idCS)
	} else { // unsolved is in O
		o, toSolve = popInternalVariable(o, idCS)
	}

	// set cS to toSolve coeff
	cS.Set(&scs.coeffs[toSolve.CoeffID()])

	// cL*cR = toSolve + cO
	f1 := func() {
		cK.Mul(&cL, &cR)
		cK.Sub(&cK, &cO)

		scs.addConstraint(compiled.SparseR1C{
			K: scs.coeffID(&cK),
			O: scs.newTerm(cS.Neg(&cS), idCS),
		})
	}

	// cL*(r + cR) = toSolve + cO
	f2 := func() {
		rt := scs.split(r)

		cRT := scs.multiply(rt, &cL)
		cK.Mul(&cL, &cR)
		cK.Sub(&cK, &cO)

		scs.addConstraint(compiled.SparseR1C{
			R: cRT,
			K: scs.coeffID(&cK),
			O: scs.newTerm(cS.Neg(&cS), idCS),
		},
		)
	}

	// (l + cL)*cR = toSolve + cO
	f3 := func() {
		lt := scs.split(l)

		cRLT := scs.multiply(lt, &cR)
		cK.Mul(&cL, &cR)
		cK.Sub(&cK, &cO)

		scs.addConstraint(compiled.SparseR1C{
			L: cRLT,
			O: scs.newTerm(cS.Neg(&cS), idCS),
			K: scs.coeffID(&cK),
		})
	}

	// (l + cL)*(r + cR) = toSolve + cO
	f4 := func() {
		lt := scs.split(l)
		rt := scs.split(r)

		cRLT := scs.multiply(lt, &cR)
		cRT := scs.multiply(rt, &cL)
		cK.Mul(&cL, &cR)
		cK.Sub(&cK, &cO)

		scs.addConstraint(compiled.SparseR1C{
			L: cRLT,
			R: cRT,
			M: [2]compiled.Term{lt, rt},
			K: scs.coeffID(&cK),
			O: scs.newTerm(cS.Neg(&cS), idCS),
		})
	}

	// cL*cR = toSolve + o + cO
	f5 := func() {
		ot := scs.split(o)

		cK.Mul(&cL, &cR)
		cK.Sub(&cK, &cO)
		cK.Neg(&cK)

		scs.addConstraint(compiled.SparseR1C{
			L: ot,
			K: scs.coeffID(&cK),
			O: scs.newTerm(cS.Neg(&cS), idCS),
		})
	}

	// cL*(r + cR) = toSolve + o + cO
	f6 := func() {
		rt := scs.split(r)
		ot := scs.split(o)

		cRT := scs.multiply(rt, &cL)
		cK.Mul(&cL, &cR)
		cK.Sub(&cK, &cO)
		cK.Neg(&cK)

		scs.addConstraint(compiled.SparseR1C{
			L: scs.negate(ot),
			R: cRT,
			K: scs.coeffID(&cK),
			O: scs.newTerm(cS.Neg(&cS), idCS),
		})
	}

	// (l + cL)*cR = toSolve + o + cO
	f7 := func() {
		lt := scs.split(l)
		ot := scs.split(o)

		cRLT := scs.multiply(lt, &cR)
		cK.Mul(&cL, &cR)
		cK.Sub(&cK, &cO)
		cK.Neg(&cK)

		scs.addConstraint(compiled.SparseR1C{
			R: scs.negate(ot),
			L: cRLT,
			K: scs.coeffID(&cK),
			O: scs.newTerm(cS.Neg(&cS), idCS),
		})
	}

	// (l + cL)*(r + cR) = toSolve + o + cO
	f8 := func() {

		lt := scs.split(l)
		rt := scs.split(r)
		ot := scs.split(o)

		cRLT := scs.multiply(lt, &cR)
		cRT := scs.multiply(rt, &cL)
		cK.Mul(&cL, &cR)
		cK.Sub(&cK, &cO)
		cK.Neg(&cK)

		u := scs.newTerm(bOne)
		scs.addConstraint(compiled.SparseR1C{
			L: cRLT,
			R: cRT,
			M: [2]compiled.Term{lt, rt},
			K: scs.coeffID(&cK),
			O: u,
		})

		scs.addConstraint(compiled.SparseR1C{
			L: u,
			R: ot,
			O: scs.newTerm(&cS, idCS),
		})
	}

	// (toSolve + cL)*cR = cO
	f9 := func() {
		cK.Mul(&cL, &cR)
		cK.Sub(&cK, &cO)

		cS.Mul(&cS, &cR)

		scs.addConstraint(compiled.SparseR1C{
			L: scs.newTerm(&cS, idCS),
			K: scs.coeffID(&cK),
		})
	}

	// (toSolve + cL)*(r + cR) = cO
	f10 := func() {
		res := scs.newTerm(&cS, idCS)

		rt := scs.split(r)
		cRT := scs.multiply(rt, &cL)
		cRes := scs.multiply(res, &cR)

		cK.Mul(&cL, &cR)
		cK.Sub(&cK, &cO)

		scs.addConstraint(compiled.SparseR1C{
			L: cRes,
			R: cRT,
			M: [2]compiled.Term{res, rt},
			K: scs.coeffID(&cK),
		})
	}

	// (toSolve + l + cL)*cR = cO
	f11 := func() {

		lt := scs.split(l)
		lt = scs.multiply(lt, &cR)

		cK.Mul(&cL, &cR)
		cK.Sub(&cK, &cO)

		cS.Mul(&cS, &cR)

		scs.addConstraint(compiled.SparseR1C{
			L: scs.newTerm(&cS, idCS),
			R: lt,
			K: scs.coeffID(&cK),
		})
	}

	// (toSolve + l + cL)*(r + cR) = cO
	// => toSolve*r + toSolve*cR + [ l*r + l*cR +cL*r+cL*cR-cO ]=0
	f12 := func() {
		u := scs.newTerm(bOne)

		lt := scs.split(l)
		rt := scs.split(r)
		cRLT := scs.multiply(lt, &cR)
		cRT := scs.multiply(rt, &cL)

		cK.Mul(&cL, &cR)
		cK.Sub(&cK, &cO)

		scs.addConstraint(compiled.SparseR1C{
			L: cRLT,
			R: cRT,
			M: [2]compiled.Term{lt, rt},
			O: u,
			K: scs.coeffID(&cK),
		})

		res := scs.newTerm(&cS, idCS)
		cRes := scs.multiply(res, &cR)

		scs.addConstraint(compiled.SparseR1C{
			R: cRes,
			M: [2]compiled.Term{res, rt},
			O: scs.negate(u),
		})
	}

	// (toSolve + cL)*cR = o + cO
	f13 := func() {
		ot := scs.split(o)

		cK.Mul(&cL, &cR)
		cK.Sub(&cK, &cO)

		cS.Mul(&cS, &cR)

		scs.addConstraint(compiled.SparseR1C{
			L: scs.newTerm(&cS, idCS),
			O: scs.negate(ot),
			K: scs.coeffID(&cK),
		})
	}

	// (toSolve + cL)*(r + cR) = o + cO
	// toSolve*r + toSolve*cR+cL*r+cL*cR-cO-o=0
	f14 := func() {
		ot := scs.split(o)
		res := scs.newTerm(&cS, idCS)

		rt := scs.split(r)

		cK.Mul(&cL, &cR)
		cK.Sub(&cK, &cO)

		scs.addConstraint(compiled.SparseR1C{
			L: scs.multiply(res, &cR),
			R: scs.multiply(rt, &cL),
			M: [2]compiled.Term{res, rt},
			O: scs.negate(ot),
			K: scs.coeffID(&cK),
		})
	}

	// (toSolve + l + cL)*cR = o + cO
	// toSolve*cR + l*cR + cL*cR-cO-o=0
	f15 := func() {

		ot := scs.split(o)
		lt := scs.split(l)

		cK.Mul(&cL, &cR)
		cK.Sub(&cK, &cO)

		cS.Mul(&cS, &cR)

		scs.addConstraint(compiled.SparseR1C{
			L: scs.newTerm(&cS, idCS),
			R: scs.multiply(lt, &cR),
			O: scs.negate(ot),
			K: scs.coeffID(&cK),
		})
	}

	// (toSolve + l + cL)*(r + cR) = o + cO
	// => toSolve*r + toSolve*cR + [ [l*r + l*cR +cL*r+cL*cR-cO]- o ]=0
	f16 := func() {
		// [l*r + l*cR +cL*r+cL*cR-cO] + u = 0
		u := scs.newTerm(bOne)

		lt := scs.split(l)
		rt := scs.split(r)
		cRLT := scs.multiply(lt, &cR)
		cRT := scs.multiply(rt, &cL)

		cK.Mul(&cL, &cR)
		cK.Sub(&cK, &cO)

		scs.addConstraint(compiled.SparseR1C{
			L: cRLT,
			R: cRT,
			M: [2]compiled.Term{lt, rt},
			O: u,
			K: scs.coeffID(&cK),
		})

		// u+o+v = 0 (v = -u - o = [l*r + l*cR +cL*r+cL*cR-cO] -  o)
		v := scs.newTerm(bOne)
		ot := scs.split(o)
		scs.addConstraint(compiled.SparseR1C{
			L: u,
			R: ot,
			O: v,
		})

		// toSolve*r + toSolve*cR + v = 0
		res := scs.newTerm(&cS, idCS)
		cRes := scs.multiply(res, &cR)

		scs.addConstraint(compiled.SparseR1C{
			R: cRes,
			M: [2]compiled.Term{res, rt},
			O: v,
		})
	}

	// we have 16 different cases
	var s uint8
	if lro != 0 {
		s |= 0b1000
	}
	if len(o) != 0 {
		s |= 0b0100
	}
	if len(l) != 0 {
		s |= 0b0010
	}
	if len(r) != 0 {
		s |= 0b0001
	}

	switch s {
	case 0b0000:
		f9()
	case 0b0001:
		f10()
	case 0b0010:
		f11()
	case 0b0011:
		f12()
	case 0b0100:
		f13()
	case 0b0101:
		f14()
	case 0b0110:
		f15()
	case 0b0111:
		f16()
	case 0b1000:
		f1()
	case 0b1001:
		f2()
	case 0b1010:
		f3()
	case 0b1011:
		f4()
	case 0b1100:
		f5()
	case 0b1101:
		f6()
	case 0b1110:
		f7()
	case 0b1111:
		f8()
	}

}

// splitR1C splits a r1c assertion (meaning that
// it's a r1c constraint that is not used to solve a variable,
// like a boolean constraint).
// (l + cL)*(r + cR) = o + cO
func (scs *sparseR1CS) splitR1C(r1c compiled.R1C) {

	l := r1c.L
	r := r1c.R
	o := r1c.O

	sort.Sort(l)
	sort.Sort(r)
	sort.Sort(o)

	l, cL := scs.popConstantTerm(l)
	r, cR := scs.popConstantTerm(r)
	o, cO := scs.popConstantTerm(o)

	var cK big.Int

	if len(o) == 0 {

		if len(l) == 0 {

			if len(r) == 0 { // cL*cR = cO (should never happen...)

				cK.Mul(&cL, &cR)
				cK.Sub(&cK, &cO)

				scs.addConstraint(compiled.SparseR1C{K: scs.coeffID(&cK)})

			} else { // cL*(r + cR) = cO

				//rt := scs.split(0, r)
				rt := scs.split(r)

				cRLT := scs.multiply(rt, &cL)
				cK.Mul(&cL, &cR)
				cK.Sub(&cK, &cO)

				scs.addConstraint(compiled.SparseR1C{R: cRLT, K: scs.coeffID(&cK)})
			}

		} else {

			if len(r) == 0 { // (l + cL)*cR = cO
				//lt := scs.split(0, l)
				lt := scs.split(l)

				cRLT := scs.multiply(lt, &cR)
				cK.Mul(&cL, &cR)
				cK.Sub(&cK, &cO)

				scs.addConstraint(compiled.SparseR1C{L: cRLT, K: scs.coeffID(&cK)})

			} else { // (l + cL)*(r + cR) = cO

				// lt := scs.split(0, l)
				// rt := scs.split(0, r)
				lt := scs.split(l)
				rt := scs.split(r)

				cRLT := scs.multiply(lt, &cR)
				cRT := scs.multiply(rt, &cL)
				cK.Mul(&cL, &cR)
				cK.Sub(&cK, &cO)

				scs.addConstraint(compiled.SparseR1C{
					L: cRLT,
					R: cRT,
					M: [2]compiled.Term{lt, rt},
					K: scs.coeffID(&cK),
				})
			}
		}

	} else {
		if len(l) == 0 {

			if len(r) == 0 { // cL*cR = o + cO

				//ot := scs.split(0, o)
				ot := scs.split(o)

				cK.Mul(&cL, &cR)
				cK.Sub(&cK, &cO)

				scs.addConstraint(compiled.SparseR1C{K: scs.coeffID(&cK), O: scs.negate(ot)})

			} else { // cL * (r + cR) = o + cO

				// rt := scs.split(0, r)
				// ot := scs.split(0, o)
				rt := scs.split(r)
				ot := scs.split(o)

				cRT := scs.multiply(rt, &cL)
				cK.Mul(&cL, &cR)
				cK.Sub(&cK, &cO)

				scs.addConstraint(compiled.SparseR1C{
					R: cRT,
					K: scs.coeffID(&cK),
					O: scs.negate(ot),
				})
			}

		} else {
			if len(r) == 0 { // (l + cL) * cR = o + cO

				// lt := scs.split(0, l)
				// ot := scs.split(0, o)
				lt := scs.split(l)
				ot := scs.split(o)

				cRLT := scs.multiply(lt, &cR)
				cK.Mul(&cL, &cR)
				cK.Sub(&cK, &cO)

				scs.addConstraint(compiled.SparseR1C{
					L: cRLT,
					K: scs.coeffID(&cK),
					O: scs.negate(ot),
				})

			} else { // (l + cL)*(r + cR) = o + cO
				// lt := scs.split(0, l)
				// rt := scs.split(0, r)
				// ot := scs.split(0, o)
				lt := scs.split(l)
				rt := scs.split(r)
				ot := scs.split(o)

				cRT := scs.multiply(rt, &cL)
				cRLT := scs.multiply(lt, &cR)
				cK.Mul(&cR, &cL)
				cK.Sub(&cK, &cO)

				scs.addConstraint(compiled.SparseR1C{
					L: cRLT,
					R: cRT,
					M: [2]compiled.Term{lt, rt},
					K: scs.coeffID(&cK),
					O: scs.negate(ot),
				})
			}
		}
	}
}

var bigIntPool = sync.Pool{
	New: func() interface{} {
		return new(big.Int)
	},
}
