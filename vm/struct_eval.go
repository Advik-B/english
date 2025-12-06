package vm

import (
	"english/ast"
	"fmt"
)

// evalStructDecl evaluates a struct declaration
func (ev *Evaluator) evalStructDecl(node *ast.StructDecl) (Value, error) {
	// Create field definitions
	fields := make(map[string]*FieldDefinition)
	fieldOrder := make([]string, 0, len(node.Fields))

	for _, field := range node.Fields {
		// Parse type name
		typeKind := ParseTypeString(field.TypeName)
		typeInfo := &TypeInfo{
			Kind: typeKind,
			Name: field.TypeName,
		}

		// Evaluate default value if provided
		var defaultValue Value
		if field.DefaultValue != nil {
			val, err := ev.Eval(field.DefaultValue)
			if err != nil {
				return nil, err
			}
			defaultValue = val
		} else {
			// Set default values based on type
			switch typeKind {
			case TypeI32:
				defaultValue = int32(0)
			case TypeI64:
				defaultValue = int64(0)
			case TypeU32:
				defaultValue = uint32(0)
			case TypeU64:
				defaultValue = uint64(0)
			case TypeF32:
				defaultValue = float32(0.0)
			case TypeF64:
				defaultValue = float64(0.0)
			case TypeString:
				defaultValue = ""
			case TypeBool:
				defaultValue = false
			case TypeList:
				defaultValue = []interface{}{}
			default:
				defaultValue = nil
			}
		}

		fields[field.Name] = &FieldDefinition{
			Name:         field.Name,
			TypeInfo:     typeInfo,
			DefaultValue: defaultValue,
		}
		fieldOrder = append(fieldOrder, field.Name)
	}

	// Create method definitions
	methods := make(map[string]*FunctionValue)
	for _, method := range node.Methods {
		// Create function value for the method
		methods[method.Name] = &FunctionValue{
			Name:       method.Name,
			Parameters: method.Parameters,
			Body:       method.Body,
			Closure:    ev.env, // Methods capture the struct definition environment
		}
	}

	// Create struct definition
	structDef := &StructDefinition{
		Name:       node.Name,
		Fields:     fields,
		Methods:    methods,
		FieldOrder: fieldOrder,
	}

	// Register struct in environment
	ev.env.DefineStruct(node.Name, structDef)

	return nil, nil
}

// evalStructInstantiation evaluates creating a new struct instance
func (ev *Evaluator) evalStructInstantiation(node *ast.StructInstantiation) (Value, error) {
	// Get struct definition
	structDef, ok := ev.env.GetStruct(node.StructName)
	if !ok {
		return nil, ev.runtimeError(fmt.Sprintf("undefined struct type '%s'", node.StructName))
	}

	// Create instance with default field values
	fields := make(map[string]Value)
	for fieldName, fieldDef := range structDef.Fields {
		fields[fieldName] = fieldDef.DefaultValue
	}

	// Override with provided field values
	for _, fieldName := range node.FieldOrder {
		expr := node.FieldValues[fieldName]
		val, err := ev.Eval(expr)
		if err != nil {
			return nil, err
		}

		// Check if field exists
		if _, ok := structDef.Fields[fieldName]; !ok {
			return nil, ev.runtimeError(fmt.Sprintf("struct '%s' has no field '%s'", node.StructName, fieldName))
		}

		fields[fieldName] = val
	}

	return &StructInstance{
		Definition: structDef,
		Fields:     fields,
	}, nil
}

// evalFieldAccess evaluates accessing a field of a struct
func (ev *Evaluator) evalFieldAccess(node *ast.FieldAccess) (Value, error) {
	// Evaluate the object expression
	obj, err := ev.Eval(node.Object)
	if err != nil {
		return nil, err
	}

	// Check if it's a struct instance
	structInst, ok := obj.(*StructInstance)
	if !ok {
		return nil, ev.runtimeError(fmt.Sprintf("cannot access field '%s' on non-struct value", node.Field))
	}

	// Get field value
	value, ok := structInst.Fields[node.Field]
	if !ok {
		return nil, ev.runtimeError(fmt.Sprintf("struct '%s' has no field '%s'", structInst.Definition.Name, node.Field))
	}

	return value, nil
}

// evalFieldAssignment evaluates assigning to a struct field
func (ev *Evaluator) evalFieldAssignment(node *ast.FieldAssignment) (Value, error) {
	// Get the struct instance
	obj, ok := ev.env.Get(node.ObjectName)
	if !ok {
		return nil, ev.runtimeError(fmt.Sprintf("undefined variable '%s'", node.ObjectName))
	}

	structInst, ok := obj.(*StructInstance)
	if !ok {
		return nil, ev.runtimeError(fmt.Sprintf("'%s' is not a struct instance", node.ObjectName))
	}

	// Check if field exists
	if _, ok := structInst.Fields[node.Field]; !ok {
		return nil, ev.runtimeError(fmt.Sprintf("struct '%s' has no field '%s'", structInst.Definition.Name, node.Field))
	}

	// Evaluate the value
	value, err := ev.Eval(node.Value)
	if err != nil {
		return nil, err
	}

	// Assign to field
	structInst.Fields[node.Field] = value

	return nil, nil
}

// evalMethodCall evaluates calling a method on an object
func (ev *Evaluator) evalMethodCall(node *ast.MethodCall) (Value, error) {
	// Evaluate the object
	obj, err := ev.Eval(node.Object)
	if err != nil {
		return nil, err
	}

	// Check if it's a struct instance
	structInst, ok := obj.(*StructInstance)
	if !ok {
		return nil, ev.runtimeError(fmt.Sprintf("cannot call method on non-struct value"))
	}

	// Get method from struct definition
	method, ok := structInst.Definition.Methods[node.MethodName]
	if !ok {
		return nil, ev.runtimeError(fmt.Sprintf("struct '%s' has no method '%s'", structInst.Definition.Name, node.MethodName))
	}

	// Evaluate arguments
	args := make([]Value, len(node.Arguments))
	for i, arg := range node.Arguments {
		val, err := ev.Eval(arg)
		if err != nil {
			return nil, err
		}
		args[i] = val
	}

	// Check parameter count
	if len(args) != len(method.Parameters) {
		return nil, ev.runtimeError(fmt.Sprintf("method '%s' expects %d arguments, got %d", node.MethodName, len(method.Parameters), len(args)))
	}

	// Create new environment for method execution
	// The method has access to struct fields as well as parameters
	methodEnv := method.Closure.NewChild()

	// Bind struct fields to method environment
	for fieldName, fieldValue := range structInst.Fields {
		methodEnv.Define(fieldName, fieldValue, false)
	}

	// Bind parameters
	for i, param := range method.Parameters {
		methodEnv.Define(param, args[i], false)
	}

	// Save current environment and switch to method environment
	oldEnv := ev.env
	ev.env = methodEnv

	// Add to call stack
	ev.callStack = append(ev.callStack, fmt.Sprintf("%s.%s()", structInst.Definition.Name, node.MethodName))

	// Execute method body
	var result Value
	for _, stmt := range method.Body {
		val, err := ev.Eval(stmt)
		if err != nil {
			// Restore environment and call stack before returning error
			ev.env = oldEnv
			ev.callStack = ev.callStack[:len(ev.callStack)-1]
			return nil, err
		}
		if retVal, ok := val.(*ReturnValue); ok {
			result = retVal.Value
			break
		}
		result = val
	}

	// Restore environment and call stack
	ev.env = oldEnv
	ev.callStack = ev.callStack[:len(ev.callStack)-1]

	// Update struct fields from method environment (in case method modified them)
	for fieldName := range structInst.Fields {
		if val, ok := methodEnv.Get(fieldName); ok {
			structInst.Fields[fieldName] = val
		}
	}

	return result, nil
}
