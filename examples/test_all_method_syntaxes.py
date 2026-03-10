# Transpiled from English language source
# Test all method call syntaxes including possessive
class Person:
    def __init__(self, name=""):
        self.name = name

    def talk(self):
        print("Hello, my name is", self.name)

    def greet(self, other):
        print(self.name, "says hello to", other)

alice = Person(name="Alice")
bob = Person(name="Bob")
# Test all three method call syntaxes
print("=== Testing possessive syntax ===")
alice.talk()
bob.talk()
print("=== Testing 'from' syntax ===")
alice.talk()
bob.talk()
print("=== Testing 'on' syntax ===")
alice.talk()
bob.talk()
print("=== Testing with arguments ===")
alice.greet("Charlie")
bob.greet("Diana")
alice.greet("Eve")
print("All method call syntaxes work!")
