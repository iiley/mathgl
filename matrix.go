package mathgl

import (
	"errors"
)

type Matrix struct {
	m, n int // an m x n matrix
	typ  VecType
	dat  []VecNum
}

func NewMatrix(m, n int, typ VecType) *Matrix {
	return &Matrix{m: m, n: n, typ: typ, dat: make([]VecNum, 0, 2)}
}

// [1, 1]
// [0, 1] Would be entered as a 2D array [[1,1],[0,1]] -- but converted to RMO
//
// This may seem confusing, but it's because it's easier to type out and visualize things in CMO
// So it's easier to type write your matrix as a slice in CMO, and pass it into this method
func MatrixFromCols(typ VecType, el [][]VecNum) (mat *Matrix, err error) {
	mat.typ = typ

	mat.m = len(el)
	mat.n = len(el[0])
	mat.dat = make([]VecNum, 0, mat.m*mat.n)

	// Row Major Order, like in OpenGL
	for i := 0; i < mat.n; i++ {
		for j := 0; j < mat.m; j++ {
			if !checkType(mat.typ, el[j][i]) {
				return nil, errors.New("Element type does not match matrix")
			}
			mat.dat = append(mat.dat, el[j][i])
		}
	}

	return mat, nil
}

// This function is MatrixOf, except it takes a list of row "vectors" instead of row "vectors" (really slices)
func MatrixFromRows(typ VecType, el [][]VecNum) (mat *Matrix, err error) {
	mat.typ = typ

	mat.m = len(el)
	mat.n = len(el[0])
	mat.dat = make([]VecNum, 0, mat.m*mat.n)

	// Row Major Order, like in OpenGL
	for i := 0; i < mat.m; i++ {
		for j := 0; j < mat.n; j++ {
			if !checkType(mat.typ, el[j][i]) {
				return nil, errors.New("Element type does not match matrix")
			}
			mat.dat = append(mat.dat, el[j][i])
		}
	}

	return mat, nil
}

// Slice-format data should be in Row Major Order
func MatrixFromSlice(typ VecType, el []VecNum, m, n int) (mat *Matrix, err error) {
	mat.typ = typ
	mat.m = m
	mat.n = n

	if mat.m*mat.n != len(el) {
		return nil, errors.New("Matrix dimensions do not match data passed in")
	}

	for _, e := range el {
		if !checkType(mat.typ, e) {
			return nil, errors.New("Type of at least one element does not match declared type")
		}
	}

	mat.dat = el

	return mat, nil
}

// Quick and dirty internal function to make a matrix without spending time checking types
func unsafeMatrixFromSlice(typ VecType, el []VecNum, m, n int) (mat *Matrix, err error) {
	mat.typ = typ
	mat.m = m
	mat.n = n

	mat.dat = el

	return mat, nil
}

// TODO: "Add" or "Append" data method (expand the matrix)

func (mat *Matrix) SetElement(i, j int, el VecNum) error {
	if i < mat.m || j < mat.n {
		return errors.New("Dimensions out of bounds")
	}

	if !checkType(mat.typ, el) {
		return errors.New("Type of element does not match matrix's type")
	}

	mat.dat[mat.m*j+i] = el

	return nil
}

func (mat Matrix) AsVector() (v Vector, err error) {
	if mat.m != 1 && mat.n != 1 {
		return v, errors.New("Matrix is not 1-dimensional in either direction.")
	}

	vPoint, err := VectorOf(mat.typ, mat.dat)
	if err != nil {
		return v, err
	}

	return *vPoint, nil
}

func (mat Matrix) ToScalar() VecNum {
	if mat.m != 1 || mat.n != 1 {
		return nil
	}

	return mat.dat[0]
}

func (m1 Matrix) Add(m2 Matrix) (m3 Matrix) {
	if m1.typ != m2.typ || len(m1.dat) != len(m2.dat) {
		return
	}

	m3.typ = m1.typ
	m3.dat = make([]VecNum, len(m1.dat))

	for i := range m1.dat {
		m3.dat[i] = m1.dat[i].add(m2.dat[i])
	}

	return m3
}

func (m1 Matrix) Sub(m2 Matrix) (m3 Matrix) {
	if m1.typ != m2.typ || len(m1.dat) != len(m2.dat) {
		return
	}

	m3.typ = m1.typ
	m3.dat = make([]VecNum, len(m1.dat))

	for i := range m1.dat {
		m3.dat[i] = m1.dat[i].sub(m2.dat[i])
	}

	return m3
}

func (m1 Matrix) Mul(m2 Matrix) (m3 Matrix) {
	if m1.n != m2.m || m1.typ != m2.typ {
		return
	}
	dat := make([]VecNum, m1.m*m2.n)

	for j := 0; j < m2.n; j++ { // Columns of m2 and m3
		for i := 0; i < m1.m; i++ { // Rows of m1 and m3
			for k := 0; k < m1.n; k++ { // Columns of m1, rows of m2
				dat[j*m2.n+i] = dat[j*m2.n+i].add(m1.dat[k*m1.n+i].mul(m2.dat[j*m2.n+k])) // I think, needs testing
			}
		}
	}

	mat, err := unsafeMatrixFromSlice(m1.typ, dat, m1.m, m2.n)
	if err != nil {
		return
	}

	return *mat
}


/* 
Batch Multiply, as its name implies. Is supposed to multiply a huge amount of matrices at once
Since matrix multiplication is associative, it can do pieces of the problem at the same time.
Since starting a goroutine has some overhead, I'd wager it's probably not worth it to use this function unless you have
6-8+ matrices, but I haven't benchmarked it so until then who knows?

Make sure the matrices are in order, and that they can indeed be multiplied. If not you'll end up with an untyped 0x0 matrix (a.k.a the "zero type" for a Matrix struct)
*/
func BatchMultiply(args []Matrix) Matrix {
	if len(args) == 1 {
		return args[0]
	}
	
	var m1,m2 Matrix
	if len(args) > 2 {
			ch1 := make(chan Matrix)
			ch2 := make(chan Matrix)
			
			// Split up the work, matrix mult is associative
			go batchMultHelper(ch1, args[0:len(args)/2])
			go batchMultHelper(ch2, args[len(args)/2:len(args)])
			
			m1 = <- ch1
			m2 = <- ch2
	} else {
		m1 = args[0]
		m2 = args[1]
	}
	
	
	return m1.Mul(m2)
}

// Wrapper so we can use multiply concurrently. Code duplication might be faster (if concurrency is faster here at all, that is). We'll need benchmarks to be sure
func batchMultHelper(ch chan<- Matrix, args[]Matrix) { 
	ch <- BatchMultiply(args)
	close(ch)
}

// INCOMPLETE DO NOT USE
// Need better function to make matrices for recursion
func (m1 Matrix) Det() interface{} {
	if m1.m != m1.n { // Determinants are only for square matrices
		return nil
	}

	return nil
}
