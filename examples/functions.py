# Transpiled from English language source
# Functions Example
# Demonstrates function declaration, parameters, and return values
print("=== Function Examples ===")
# Simple function with no parameters
def sayHello():
    print("Hello from the function!")

print("Calling sayHello:")
sayHello()
# Function with parameters
def greet(name):
    print("Hello,", name, "! Nice to meet you.")

print("Calling greet with 'Alice':")
x = 0
x = greet("Alice")
# Function with return value
def add(a, b):
    return a + b

print("Adding 5 and 7:")
sum = 0
sum = add(5, 7)
print("Result:", sum)
# Function with multiple parameters
def multiply(x, y):
    return x * y

print("Multiplying 6 and 8:")
product = 0
product = multiply(6, 8)
print("Result:", product)
# Nested function calls
print("Nested: add(multiply(2, 3), 4)")
inner = 0
inner = multiply(2, 3)
outer = 0
outer = add(inner, 4)
print("Result:", outer)
# Function that uses conditionals
def max(a, b):
    if a > b:
        return a
    else:
        return b

print("Max of 10 and 25:")
maximum = 0
maximum = max(10, 25)
print("Result:", maximum)
# Function that calculates average
def average(x, y, z):
    total = (x + y) + z
    return total / 3

print("Average of 10, 20, 30:")
avg = 0
avg = average(10, 20, 30)
print("Result:", avg)
print("=== Done! ===")
