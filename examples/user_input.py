# Transpiled from English language source
# User Input Example
# Demonstrates reading input from the user
print("=== User Input Demo ===")
print("This program asks for your name and age, then greets you.")
print("")
# Ask for user input using the Ask statement
name = input("What is your name? ")
age_str = input("How old are you? ")
age = float(age_str)
print("")
print("Hello, ")
print(name)
next_year = age + 1
print("Next year you will be ")
print(next_year)
if age >= 18:
    print("You are an adult.")
else:
    print("You are a minor.")
print("=== Done! ===")
