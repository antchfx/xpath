test xml document
====

```xml
<?xml version="1.0" encoding="UTF-8"?>
<bookstore>
<book>
  <title lang="en" id="1">Harry Potter</title>
  <price>29.99</price>
</book>
<book>
  <title lang="en" id="2">Learning XML</title>
  <price>39.95</price>
</book>
</bookstore>
```

How to run test example
===
1. go run main.go

2. enter your XPath expression.(etc,//book,//@lang,//book[price>20])
