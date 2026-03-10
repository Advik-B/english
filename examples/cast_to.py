# Transpiled from English language source
# Cast To Example
# Demonstrates the natural English "cast to" type conversion syntax
print("=== Cast To Syntax ===")
# String to number
age_str = "25"
age = float(age_str)
print("String '25' cast to number:")
print(age)
print("Age plus 5:")
print(age + 5)
print("")
# Number to text
score = 98
score_text = str(score)
print("Score as text:")
print(score_text)
print("Score starts with '9':")
print(score_text.startswith("9"))
print("")
# Float string to number
pi_str = "3.14159"
pi_approx = float(pi_str)
print("Pi approximation:")
print(pi_approx)
print("")
# Using "casted" (alternative keyword)
temp_str = "98.6"
temp = float(temp_str)
print("Temperature:")
print(temp)
print("")
# Cast in conditions
print("=== Cast in Conditions ===")
threshold_str = "50"
score2 = 75
if score2 > float(threshold_str):
    print("Value exceeds threshold.")
# Boolean cast
print("")
print("=== Boolean Cast ===")
zero = 0
one = 1
if not bool(zero):
    print("0 cast to boolean is false.")
if bool(one):
    print("1 cast to boolean is true.")
print("=== Done! ===")
