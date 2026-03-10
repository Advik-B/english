# Transpiled from English language source
# String Functions Example
# Demonstrates the extended string standard library
print("=== Extended String Functions ===")
text = "Hello, World! Hello, English!"
print("Original text:")
print(text)
print("")
# Checking prefixes and suffixes
print("Starts with 'Hello':")
print(text.startswith("Hello"))
print("Ends with 'English!':")
print(text.endswith("English!"))
# Finding substrings
print("Index of 'World':")
print(text.find("World"))
print("Index of 'missing':")
print(text.find("missing"))
# Extracting substrings
print("Substring from index 7, length 5:")
print(text[7:7+5])
# Repeating strings
print("Repeat 'ha' 5 times:")
print("ha" * 5)
# Counting occurrences
print("Count of 'Hello':")
print(text.count("Hello"))
print("Count of 'l':")
print(text.count("l"))
# Padding
print("Pad '42' to width 8 with '0' on left:")
print("42".rjust(8, "0"))
print("Pad 'hi' to width 10 with '.' on right:")
print("hi".ljust(10, "."))
# Type Conversions
print("")
print("=== Type Conversions ===")
num_str = "3.14159"
as_num = float(num_str)
print("String '3.14159' cast to number:")
print(as_num)
print("Number 42 cast to text:")
print(str(42))
# Checking empty
print("")
print("=== Empty Checks ===")
print("Is '' empty:")
print((len("") == 0))
print("Is 'hello' empty:")
print((len("hello") == 0))
print("Is [] empty:")
print((len([]) == 0))
print("=== Done! ===")
