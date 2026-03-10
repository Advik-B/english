# Transpiled from English language source
# Possessive method call test
class Person:
    def __init__(self, name=""):
        self.name = name

    def talk(self):
        print("Hello, my name is", self.name)

p = Person(name="Alice")
p.talk()
