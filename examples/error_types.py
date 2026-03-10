# Transpiled from English language source
# error_types.abc — demonstrates custom error types and static type annotations.
print("=== Custom Error Types ===")
print("")
# Declare custom error types
class NetworkError(Exception): pass

class ValidationError(Exception): pass

class DatabaseError(Exception): pass

# Raise and catch by type
try:
    raise NetworkError("Host unreachable")
except NetworkError as error:
    print("Handled network error:", error)
finally:
    print("Network cleanup done.")
# Catch-all still works for any error
try:
    raise ValidationError("Age must be positive")
except Exception as error:
    print("Caught any error:", error)
# Type-specific catch: wrong type propagates to outer handler
try:
    try:
        raise DatabaseError("Row not found")
    except NetworkError as error:
        print("This should NOT print.")
except DatabaseError as error:
    print("Outer caught database error:", error)
print("")
print("=== Static Type Annotations ===")
print("")
# Declare variables with explicit types
count: float = 0
greeting: str = "Hello"
active: bool = True
print("count:", count)
print("greeting:", greeting)
print("active:", active)
# Typed variable with no initial value
score: float = None
score = 95
print("score:", score)
# Type enforcement: assigning wrong type is a TypeError
try:
    count = "not a number"
except Exception as error:
    print("Caught type error:", error)
print("")
print("All tests passed!")
