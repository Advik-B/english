# Transpiled from English language source
# Test struct methods and method call syntax
class Person:
    def __init__(self, name="", age=18):
        self.name = name
        self.age = age

    def talk(self):
        print("Hello, my name is", self.name)

    def greet(self, other):
        print(self.name, "says hello to", other)

# Create instance
p = Person(name="Alice", age=25)
# Test method call with "from" (preferred)
print("Testing 'call from' syntax:")
p.talk()
# Test method call with "on"
print("Testing 'call on' syntax:")
p.talk()
# Test method with arguments
print("Testing method with arguments:")
p.greet("Bob")
print("All method call syntaxes work!")
