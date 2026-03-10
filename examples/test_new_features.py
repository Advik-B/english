# Transpiled from English language source
# Test struct and type system features
# Define a Person struct
class Person:
    def __init__(self, name="", age=18):
        self.name = name
        self.age = age

# Create an instance with default values
p1 = Person()
print("Created p1 with defaults")
# Create an instance with specific values
p2 = Person(name="John Doe", age=25)
print("Created p2 with name:", p2.name, "and age:", p2.age)
# Test type expressions
x = 10
print("Type of x:", type(x).__name__)
word = "hello"
print("Type of word:", type(word).__name__)
# Test swap
a = 1
b = 2
print("Before swap: a =", a, "b =", b)
a, b = b, a
print("After swap: a =", a, "b =", b)
# Test error handling
try:
    print("About to cause an error...")
    raise RuntimeError("This is a test error")
    print("This should not print")
except Exception as error:
    print("Caught error:", error)
finally:
    print("Finally block executed")
print("Program completed successfully!")
