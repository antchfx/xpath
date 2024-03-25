package xpath

import (
	"testing"
)

func TestLogicals(t *testing.T) {
	test_xpath_elements(t, book_example, `//book[1 + 1]`, 9)
	test_xpath_elements(t, book_example, `//book[1 * 2]`, 9)
	test_xpath_elements(t, book_example, `//book[5 div 2]`, 9) // equal to `//book[2]`
	test_xpath_elements(t, book_example, `//book[3 div 2]`, 3)
	test_xpath_elements(t, book_example, `//book[3 - 2]`, 3)
	test_xpath_elements(t, book_example, `//book[price > 35]`, 15, 25)
	test_xpath_elements(t, book_example, `//book[price >= 30]`, 3, 15, 25)
	test_xpath_elements(t, book_example, `//book[price < 30]`, 9)
	test_xpath_elements(t, book_example, `//book[price <= 30]`, 3, 9)
	test_xpath_elements(t, book_example, `//book[count(author) > 1]`, 15)
	test_xpath_elements(t, book_example, `//book[position() mod 2 = 0]`, 9, 25)
}

func TestPositions(t *testing.T) {
	test_xpath_elements(t, employee_example, `/empinfo/employee[2]`, 8)
	test_xpath_elements(t, employee_example, `//employee[position() = 2]`, 8)
	test_xpath_elements(t, employee_example, `/empinfo/employee[2]/name`, 9)
	test_xpath_elements(t, employee_example, `//employee[position() > 1]`, 8, 13)
	test_xpath_elements(t, employee_example, `//employee[position() <= 2]`, 3, 8)
	test_xpath_elements(t, employee_example, `//employee[last()]`, 13)
	test_xpath_elements(t, employee_example, `//employee[position() = last()]`, 13)
	test_xpath_elements(t, book_example, `//book[@category = "web"][2]`, 25)
	test_xpath_elements(t, book_example, `(//book[@category = "web"])[2]`, 25)
}

func TestPredicates(t *testing.T) {
	test_xpath_elements(t, employee_example, `//employee[name]`, 3, 8, 13)
	test_xpath_elements(t, employee_example, `/empinfo/employee[@id]`, 3, 8, 13)
	test_xpath_elements(t, book_example, `//book[@category = "web"]`, 15, 25)
	test_xpath_elements(t, book_example, `//book[author = "J K. Rowling"]`, 9)
	test_xpath_elements(t, book_example, `//book[./author/text() = "J K. Rowling"]`, 9)
	test_xpath_elements(t, book_example, `//book[year = 2005]`, 3, 9)
	test_xpath_elements(t, book_example, `//year[text() = 2005]`, 6, 12)
	test_xpath_elements(t, employee_example, `/empinfo/employee[1][@id=1]`, 3)
	test_xpath_elements(t, employee_example, `/empinfo/employee[@id][2]`, 8)
}

func TestOperators(t *testing.T) {
	test_xpath_elements(t, employee_example, `//designation[@discipline and @experience]`, 5, 10)
	test_xpath_elements(t, employee_example, `//designation[@discipline or @experience]`, 5, 10, 15)
	test_xpath_elements(t, employee_example, `//designation[@discipline | @experience]`, 5, 10, 15)
	test_xpath_elements(t, employee_example, `/empinfo/employee[@id != "2" ]`, 3, 13)
	test_xpath_elements(t, employee_example, `/empinfo/employee[@id and @id = "2"]`, 8)
	test_xpath_elements(t, employee_example, `/empinfo/employee[@id = "1" or @id = "2"]`, 3, 8)
}

func TestNestedPredicates(t *testing.T) {
	test_xpath_elements(t, employee_example, `//employee[./name[@from]]`, 8)
	test_xpath_elements(t, employee_example, `//employee[.//name[@from = "CA"]]`, 8)
}
