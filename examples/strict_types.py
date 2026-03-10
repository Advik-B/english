# Transpiled from English language source
# strict_types.abc - demonstrates the strict static type system
print("=== Strict Static Type System ===")
print("")
# --- Type inference at declaration ---
print("Types are inferred at declaration and fixed forever.")
counter = 0
greeting = "Hello"
active = True
# --- Valid reassignment (same type) ---
counter = counter + 1
greeting = "Hi there"
active = False
print("Valid reassignments work fine.")
# --- Explicit cast to convert types ---
print("")
print("=== Explicit Cast ===")
age_str = "25"
age = float(age_str)
if age > 18:
    print("Adult (cast worked).")
score = 97
score_label = str(score)
print(score_label.startswith("9"))
# --- Strict arithmetic (no mixing types) ---
print("")
print("=== Strict Arithmetic ===")
a = 10
b = 3
print(a + b)
print(a - b)
print(a * b)
# --- Strict text concatenation ---
print("")
print("=== Strict Text Concatenation ===")
first_name = "Jane"
last_name = "Doe"
full_name = (first_name + " ") + last_name
print(full_name)
# --- Boolean conditions must be boolean ---
print("")
print("=== Boolean Conditions ===")
x = 5
if x > 3:
    print("Comparison returns boolean: OK.")
done = False
if not done:
    print("Logical 'not' on boolean: OK.")
# --- nothing is universal null ---
print("")
print("=== Nothing (null) ===")
result = None
if result == None:
    print("Unset result is nothing.")
print("=== Done ===")
