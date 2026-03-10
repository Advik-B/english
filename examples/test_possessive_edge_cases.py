# Transpiled from English language source
# Test possessive syntax doesn't interfere with other features
class Book:
    def __init__(self, title="", author=""):
        self.title = title
        self.author = author

    def describe(self):
        print("Book:", self.title, "by", self.author)

# Create instance
mybook = Book(title="The Great Gatsby", author="F. Scott Fitzgerald")
# Test field access (no possessive)
print("Title:", mybook.title)
print("Author:", mybook.author)
# Test possessive method call
print("Using possessive syntax:")
mybook.describe()
# Test that regular variables with 's' in the name still work
persons = [1, 2, 3]
print("List persons:", persons)
# Test identifiers ending with 's'
items = 5
class_ = "A"
print("items:", items, "class:", class_)
print("All tests passed!")
