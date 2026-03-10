# Transpiled from English language source
# nil_check.abc - demonstrates "is something" / "has a value" nil-check syntax
print("=== Nil-Check Expressions ===")
print("")
# - Basic value check -
name = "Alice"
if name is not None:
    print("name has a value.")
if name is not None:
    print("name has a value (alternate syntax).")
# - Nothing literal -
result = None
if result is None:
    print("result is nothing.")
if result is None:
    print("result has no value (alternate syntax).")
# - The "is nothing" form as a postfix operator -
if result is not None:
    print("result is set.")
else:
    print("result is not yet set.")
# - Clear a variable -
print("")
score = 95
print("Before clear:")
if score is not None:
    print("score is set.")
score = None
print("After clear:")
if score is not None:
    print("score is still set.")
else:
    print("score was cleared.")
# - Guard pattern -
print("")
print("=== Guard Pattern ===")
user = None


def greet(username):
    if username is not None:
        print("Hello, ")
        print(username)
    else:
        print("Hello, stranger.")

greet("Bob")
greet(None)
# - Works with all types -
print("")
print("=== All Types ===")
n = 42
t = "hello"
b = True
empty_n = None
if n is not None:
    print("number is something.")
if t is not None:
    print("text is something.")
if b is not None:
    print("boolean is something.")
if empty_n is not None:
    print("empty is something.")
else:
    print("empty is nothing.")
print("")
print("=== Done ===")
