package ivm

// Opcode is a single-byte instruction code.
type Opcode byte

const (
	// ── Constants ─────────────────────────────────────────────────────────
	OP_LOAD_CONST   Opcode = iota // push constants[operand]
	OP_LOAD_NOTHING               // push nil

	// ── Variables ─────────────────────────────────────────────────────────
	OP_LOAD_VAR        // push env[names[operand]]
	OP_STORE_VAR       // env.Set(names[operand], pop())
	OP_DEFINE_VAR      // env.Define(names[operand], pop(), false)
	OP_DEFINE_CONST    // env.Define(names[operand], pop(), true)
	OP_DEFINE_TYPED    // pop value, pop type_name_str; env.DefineTyped(names[operand], type, val, false)
	OP_DEFINE_TYPED_CONST // pop value, pop type_name_str; env.DefineTyped(names[operand], type, val, true)
	OP_TOGGLE_VAR      // toggle boolean at names[operand]

	// ── Arithmetic / comparison ────────────────────────────────────────────
	OP_BINARY_OP // binary operation; operand encodes BinOp
	OP_UNARY_OP  // unary operation; operand encodes UnaryOp

	// ── Control flow ──────────────────────────────────────────────────────
	OP_JUMP          // jump to operand (absolute)
	OP_JUMP_IF_FALSE // pop condition; jump to operand if false/error
	OP_JUMP_IF_TRUE  // pop condition; jump to operand if true

	// ── Scope ─────────────────────────────────────────────────────────────
	OP_PUSH_SCOPE // push a new child environment
	OP_POP_SCOPE  // restore parent environment

	// ── Functions ─────────────────────────────────────────────────────────
	OP_DEFINE_FUNC  // define function; operand = func chunk index in chunk.Funcs
	OP_CALL         // call function; operand = argc<<16 | name_idx
	OP_CALL_METHOD  // call method; operand = argc<<16 | method_name_idx; object on stack below args
	OP_RETURN       // return top of stack (or nil if stack empty)

	// ── Output ────────────────────────────────────────────────────────────
	OP_PRINT // print; operand = count<<1 | newline_flag

	// ── Collections ───────────────────────────────────────────────────────
	OP_BUILD_LIST   // build list; operand = element count
	OP_BUILD_ARRAY  // build typed array; operand = element count; pop type name string after elements
	OP_BUILD_LOOKUP // push empty lookup table
	OP_INDEX_GET    // pop index, pop list; push list[index]
	OP_INDEX_SET    // operand = list name index; pop value, pop index; list[index] = value
	OP_LENGTH       // pop value; push length

	// ── Lookup table ──────────────────────────────────────────────────────
	OP_LOOKUP_GET // pop key, pop table; push table[key]
	OP_LOOKUP_SET // operand = table name index; pop value, pop key; table[key] = value
	OP_LOOKUP_HAS // pop key, pop table; push bool (key in table)

	// ── Type operations ───────────────────────────────────────────────────
	OP_TYPEOF          // pop value; push type name string
	OP_CAST            // operand = type name index; pop value; push cast(value, type)
	OP_NIL_CHECK       // operand: 1=is_something, 0=is_nothing; pop value; push bool
	OP_ERROR_TYPE_CHECK // operand = type name index; pop value; push bool

	// ── Input ─────────────────────────────────────────────────────────────
	OP_ASK // operand: 1=has_prompt, 0=no_prompt; [pop prompt;] push input line

	// ── Location ──────────────────────────────────────────────────────────
	OP_LOCATION // operand = name index; push address string

	// ── Struct ────────────────────────────────────────────────────────────
	OP_DEFINE_STRUCT // operand = struct def index in chunk.StructDefs
	OP_NEW_STRUCT    // operand = field_count<<16 | struct_name_idx; pop field values; push struct instance
	OP_GET_FIELD     // operand = field name index; pop object; push field value
	OP_SET_FIELD     // operand = field_name_idx; pop value, then load object by name (object_name in next operand via names); simpler: pop value, pop object; set field

	// ── Error handling ────────────────────────────────────────────────────
	OP_RAISE         // operand = type_name_idx (0 = generic/RuntimeError); pop message
	OP_TRY_BEGIN     // operand = catch offset; push try frame
	OP_TRY_END       // pop try frame; operand = end offset (jump past catch+finally)
	OP_CATCH         // operand = error_var_name_idx; bind error var (type check moved to handleError)

	// OP_TRY_SET_ERRORTYPE sets the error-type filter on the top try frame.
	// operand = nameIdx+1 (0 means catch-all / no filter).
	// Emitted immediately after OP_TRY_BEGIN when the catch clause has a type filter.
	OP_TRY_SET_ERRORTYPE

	// OP_TRY_SET_FINALLY records the bytecode offset where the finally body starts.
	// operand = finally_offset. Emitted after OP_TRY_BEGIN (and optional OP_TRY_SET_ERRORTYPE).
	// When set, handleError will jump to this offset (instead of the catch handler) on a type
	// mismatch, run the finally body, and then re-raise via OP_RERAISE_PENDING.
	OP_TRY_SET_FINALLY

	// OP_RERAISE_PENDING re-raises frame.pendingError if it is set.
	// Emitted at the end of every finally body so that a type-mismatch error gets
	// re-propagated after the finally block finishes.
	OP_RERAISE_PENDING

	// ── Error type declaration ────────────────────────────────────────────
	OP_DEFINE_ERROR_TYPE // operand = name_idx<<16 | parent_name_idx (0 = no parent)

	// ── Reference / copy ──────────────────────────────────────────────────
	OP_MAKE_REFERENCE // operand = name index; push reference value
	OP_MAKE_COPY      // pop value; push deep copy

	// ── Swap ──────────────────────────────────────────────────────────────
	OP_SWAP_VARS // operand = name1_idx<<16 | name2_idx

	// ── Import ────────────────────────────────────────────────────────────
	OP_IMPORT // operand = flags (importAll<<2 | isSafe<<1 | hasItems); top of stack = path string

	// ── Line tracking ─────────────────────────────────────────────────────
	OP_SET_LINE // operand = line number

	// ── Stack management ──────────────────────────────────────────────────
	OP_POP // discard top of stack
)

// BinOp encodes a binary operator.
type BinOp uint32

const (
	BinAdd BinOp = iota
	BinSub
	BinMul
	BinDiv
	BinMod
	BinEq
	BinNeq
	BinLt
	BinLte
	BinGt
	BinGte
)

// UnaryOp encodes a unary operator.
type UnaryOp uint32

const (
	UnaryNeg UnaryOp = iota // arithmetic negation
	UnaryNot                // logical not
)

// OpName returns a human-readable name for an opcode.
func OpName(op Opcode) string {
	switch op {
	case OP_LOAD_CONST:
		return "LOAD_CONST"
	case OP_LOAD_NOTHING:
		return "LOAD_NOTHING"
	case OP_LOAD_VAR:
		return "LOAD_VAR"
	case OP_STORE_VAR:
		return "STORE_VAR"
	case OP_DEFINE_VAR:
		return "DEFINE_VAR"
	case OP_DEFINE_CONST:
		return "DEFINE_CONST"
	case OP_DEFINE_TYPED:
		return "DEFINE_TYPED"
	case OP_DEFINE_TYPED_CONST:
		return "DEFINE_TYPED_CONST"
	case OP_TOGGLE_VAR:
		return "TOGGLE_VAR"
	case OP_BINARY_OP:
		return "BINARY_OP"
	case OP_UNARY_OP:
		return "UNARY_OP"
	case OP_JUMP:
		return "JUMP"
	case OP_JUMP_IF_FALSE:
		return "JUMP_IF_FALSE"
	case OP_JUMP_IF_TRUE:
		return "JUMP_IF_TRUE"
	case OP_PUSH_SCOPE:
		return "PUSH_SCOPE"
	case OP_POP_SCOPE:
		return "POP_SCOPE"
	case OP_DEFINE_FUNC:
		return "DEFINE_FUNC"
	case OP_CALL:
		return "CALL"
	case OP_CALL_METHOD:
		return "CALL_METHOD"
	case OP_RETURN:
		return "RETURN"
	case OP_PRINT:
		return "PRINT"
	case OP_BUILD_LIST:
		return "BUILD_LIST"
	case OP_BUILD_ARRAY:
		return "BUILD_ARRAY"
	case OP_BUILD_LOOKUP:
		return "BUILD_LOOKUP"
	case OP_INDEX_GET:
		return "INDEX_GET"
	case OP_INDEX_SET:
		return "INDEX_SET"
	case OP_LENGTH:
		return "LENGTH"
	case OP_LOOKUP_GET:
		return "LOOKUP_GET"
	case OP_LOOKUP_SET:
		return "LOOKUP_SET"
	case OP_LOOKUP_HAS:
		return "LOOKUP_HAS"
	case OP_TYPEOF:
		return "TYPEOF"
	case OP_CAST:
		return "CAST"
	case OP_NIL_CHECK:
		return "NIL_CHECK"
	case OP_ERROR_TYPE_CHECK:
		return "ERROR_TYPE_CHECK"
	case OP_ASK:
		return "ASK"
	case OP_LOCATION:
		return "LOCATION"
	case OP_DEFINE_STRUCT:
		return "DEFINE_STRUCT"
	case OP_NEW_STRUCT:
		return "NEW_STRUCT"
	case OP_GET_FIELD:
		return "GET_FIELD"
	case OP_SET_FIELD:
		return "SET_FIELD"
	case OP_RAISE:
		return "RAISE"
	case OP_TRY_BEGIN:
		return "TRY_BEGIN"
	case OP_TRY_END:
		return "TRY_END"
	case OP_CATCH:
		return "CATCH"
	case OP_TRY_SET_ERRORTYPE:
		return "TRY_SET_ERRORTYPE"
	case OP_TRY_SET_FINALLY:
		return "TRY_SET_FINALLY"
	case OP_RERAISE_PENDING:
		return "RERAISE_PENDING"
	case OP_DEFINE_ERROR_TYPE:
		return "DEFINE_ERROR_TYPE"
	case OP_MAKE_REFERENCE:
		return "MAKE_REFERENCE"
	case OP_MAKE_COPY:
		return "MAKE_COPY"
	case OP_SWAP_VARS:
		return "SWAP_VARS"
	case OP_IMPORT:
		return "IMPORT"
	case OP_SET_LINE:
		return "SET_LINE"
	case OP_POP:
		return "POP"
	default:
		return "UNKNOWN"
	}
}
