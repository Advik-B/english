# Transpiled from English language source
# Conditionals Example
# Demonstrates if/else/else-if statements
print("=== Conditionals ===")
# Simple if-else
x = 15
print("Testing with x =", x)
if x > 10:
    print("x is greater than 10")
else:
    print("x is not greater than 10")
# Else-if chain
print("Testing grade classification:")
score = 85
print("Score:", score)
if score >= 90:
    print("Grade: A")
elif score >= 80:
    print("Grade: B")
elif score >= 70:
    print("Grade: C")
elif score >= 60:
    print("Grade: D")
else:
    print("Grade: F")
# Nested conditionals
print("Testing nested conditions:")
age = 25
hasLicense = True
if age >= 18:
    print("Adult")
    if hasLicense == True:
        print("Can drive")
    else:
        print("Cannot drive without license")
else:
    print("Minor - cannot drive")
# Equality and inequality checks
a = 5
b = 5
if a == b:
    print("a equals b")
if a != 10:
    print("a is not 10")
print("=== Done! ===")
