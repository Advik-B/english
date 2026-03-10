# Transpiled from English language source
from typing import Final

# This is a comment
x = 5
# For variable declaration
y: Final = 10
# For constant declaration (can also use: to always be)
z: Final = "hello"


# Alternative constant syntax
# For function declaration
def say_hello():
    print("Hello, World!")

x = 15
# For assignment
say_hello()


# For function call
# For "return" statements inside functions
def add(a, b):
    return a + b

result = add(5, 10)
# Loops
for _ in range(5):
    print("This will be printed 5 times.")
# While loops
while x < 20:
    print("x is still less than 20.")
    x = x + 1
# If statements
if x == 20:
    print("x is now 20!")
else:
    print("x is not 20 yet.")
# Else if statements
if x < 10:
    print("x is less than 10.")
elif x < 20:
    print("x is between 10 and 19.")
else:
    print("x is 20 or more.")
# for-each loops
myList = [1, 2, 3, 4, 5]
for item in myList:
    print(item)
