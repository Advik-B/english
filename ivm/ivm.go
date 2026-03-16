// Package ivm implements a stack-based instruction VM for the English language.
// It compiles AST programs to flat instruction sequences (chunks) and executes
// them with a register-free, stack-based dispatch loop.
//
// Compared with the tree-walk evaluator in the vm package, the instruction VM:
//   - Avoids the overhead of virtual method dispatch on each AST node
//   - Produces compact binary bytecode that is faster to load than serialised AST
//   - Makes future optimisations (constant folding, dead-code elimination) easier
//
// File layout:
//
//	ivm.go       – package doc, public entry points (Compile, Execute)
//	opcode.go    – Opcode type and opcode constants
//	chunk.go     – Chunk (instruction stream + constant pool + name pool)
//	compiler.go  – Compiler: walks the AST and emits instructions
//	machine.go   – Machine: runs a Chunk using a value stack
//	encoding.go  – Binary serialisation / deserialisation of Chunks
package ivm

import (
	"github.com/Advik-B/english/ast"
	"github.com/Advik-B/english/parser"
	"os"
)

// Execute runs a compiled Chunk and returns the last value (or nil).
// builtin is the stdlib function dispatcher.
// predefined is a map of pre-defined constant values (e.g. math.Pi).
func Execute(chunk *Chunk, builtin BuiltinFunc, predefined map[string]interface{}) (interface{}, error) {
	m := newMachine(builtin)

	root := newIvmEnv()
	// Install predefined constants
	for k, v := range predefined {
		root.defineVar(k, v, true)
	}

	// Set up import handler that reads, compiles, and executes source files
	m.importHandler = func(path string, items []interface{}, importAll, isSafe bool, env *ivmEnv) error {
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		lexer := parser.NewLexer(string(src))
		tokens := lexer.TokenizeAll()
		p := parser.NewParser(tokens)
		prog, err := p.Parse()
		if err != nil {
			return err
		}

		if isSafe {
			// Safe import: execute only declarations (VariableDecl, FunctionDecl,
			// StructDecl, typed declarations, error type declarations) in the current
			// environment, skipping any top-level side-effectful statements (Print, Call,
			// etc.). This mirrors the tree-walk evalSafeImport behaviour.
			safeDecls := safeDeclsOnly(prog)
			subChunk, err := Compile(safeDecls)
			if err != nil {
				return err
			}
			subMachine := newMachine(builtin)
			subMachine.importHandler = m.importHandler
			subMachine.cur = &callFrame{
				chunk: subChunk,
				ip:    0,
				stack: []interface{}{},
				env:   env, // define directly in caller's env
			}
			_, execErr := subMachine.execute(env)
			return execErr
		}

		subChunk, err := Compile(prog)
		if err != nil {
			return err
		}
		subEnv := env.newChild()
		subMachine := newMachine(builtin)
		subMachine.importHandler = m.importHandler
		subMachine.cur = &callFrame{
			chunk: subChunk,
			ip:    0,
			stack: []interface{}{},
			env:   subEnv,
		}
		_, execErr := subMachine.execute(subEnv)
		if execErr != nil {
			return execErr
		}
		// Import specific items or all
		if !importAll && len(items) > 0 {
			for _, item := range items {
				name, _ := item.(string)
				if name == "" {
					continue
				}
				val, ok := subEnv.getVar(name)
				if ok {
					env.vars[name] = &envEntry{value: val}
				}
				fn, ok := subEnv.getFunc(name)
				if ok {
					env.funcs[name] = fn
				}
			}
		} else {
			// Import everything (importAll == true or no explicit items list)
			for k, v := range subEnv.vars {
				env.vars[k] = v
			}
			for k, v := range subEnv.funcs {
				env.funcs[k] = v
			}
			for k, v := range subEnv.structDefs {
				env.structDefs[k] = v
			}
			for k, v := range subEnv.errorTypes {
				env.errorTypes[k] = v
			}
		}
		return nil
	}

	m.cur = &callFrame{
		chunk: chunk,
		ip:    0,
		stack: []interface{}{},
		env:   root,
	}
	return m.execute(root)
}

// compileProgram is a helper used internally.
func compileProgram(prog *ast.Program) (*Chunk, error) {
	return Compile(prog)
}

// safeDeclsOnly returns a new Program containing only declaration statements
// (VariableDecl, TypedVariableDecl, FunctionDecl, StructDecl, ErrorTypeDecl).
// Used by the safe-import path to skip side-effectful top-level code (Print,
// Call, etc.), matching the tree-walk evalSafeImport.
func safeDeclsOnly(prog *ast.Program) *ast.Program {
	filtered := &ast.Program{}
	for _, stmt := range prog.Statements {
		switch stmt.(type) {
		case *ast.VariableDecl,
			*ast.TypedVariableDecl,
			*ast.FunctionDecl,
			*ast.StructDecl,
			*ast.ErrorTypeDecl:
			filtered.Statements = append(filtered.Statements, stmt)
		}
	}
	return filtered
}
