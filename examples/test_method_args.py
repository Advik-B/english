# Transpiled from English language source
# Method call with arguments test
class Person:
    def __init__(self, name=""):
        self.name = name

    def greet(self, other):
        print(self.name, "says hello to", other)

p = Person(name="Alice")
p.greet("Bob")
