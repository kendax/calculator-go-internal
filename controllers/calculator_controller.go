package controllers

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

// Initialize a session store
var store = sessions.NewCookieStore([]byte("mysession"))

// Operator precedence for basic arithmetic operations
var precedence = map[string]int{
	"+":  1,
	"-":  1,
	"*":  2,
	"/":  2,
	"*-": 3,
}

// isOperator checks if a given token is an operator
func isOperator(token string) bool {
	_, ok := precedence[token]
	return ok || (token == "-" && len(token) == 1)
}

//  performOperation applies the operator to two operands
func performOperation(operatorStack []string, valueStack []float64) ([]string, []float64) {
	operator := operatorStack[len(operatorStack)-1]

	//Remove the last element from this slice by slicing it
	operatorStack = operatorStack[:len(operatorStack)-1]

	operand2 := valueStack[len(valueStack)-1]
	valueStack = valueStack[:len(valueStack)-1]

	operand1 := valueStack[len(valueStack)-1]
	valueStack = valueStack[:len(valueStack)-1]

	switch operator {
	case "*-":
		valueStack = append(valueStack, operand1*(operand2*-1))
	case "+":
		valueStack = append(valueStack, operand1+operand2)
	case "-":
		valueStack = append(valueStack, operand1-operand2)
	case "*":
		valueStack = append(valueStack, operand1*operand2)
	case "/":
		valueStack = append(valueStack, operand1/operand2)
	}

	return operatorStack, valueStack
}

// CalculateInput evaluates a mathematical expression
func CalculateInput(expression string) (float64, error) {
	//Split the string at each place there is a whitespace
	tokens := strings.Fields(expression)

	operatorStack := []string{}
	valueStack := []float64{}

	for i, token := range tokens {
		if token == "(" {
			operatorStack = append(operatorStack, token)
		} else if token == ")" {
			for len(operatorStack) > 0 && operatorStack[len(operatorStack)-1] != "(" {
				operatorStack, valueStack = performOperation(operatorStack, valueStack)
			}
			operatorStack = operatorStack[:len(operatorStack)-1] // Pop "("
		} else if isOperator(token) {
			if token == "-" && (i == 0 || isOperator(tokens[i-1]) || tokens[i-1] == "(") { // Handle negative sign as part of the number
				valueStack = append(valueStack, -1)
				operatorStack = append(operatorStack, "*")
			} else {
				for len(operatorStack) > 0 && precedence[operatorStack[len(operatorStack)-1]] >= precedence[token] {
					operatorStack, valueStack = performOperation(operatorStack, valueStack)
				}
				operatorStack = append(operatorStack, token)
			}
		} else {
			value, err := parseNumber(token)
			if err != nil {
				return 0, err
			}
			valueStack = append(valueStack, value)
		}
	}

	for len(operatorStack) > 0 { //For any remaining values, evaluate them
		operatorStack, valueStack = performOperation(operatorStack, valueStack)
	}

	if len(valueStack) == 1 { //If only one number is submitted, return it 
		return valueStack[0], nil
	}

	return 0, fmt.Errorf("Invalid expression")
}

// parseNumber converts a string to a float64
func parseNumber(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

//Add a whitespace at every occurrence of an operator in a string so that the string can be further split by "strings.Fields"
func addWhitespaceAroundOperators(input string) string {
	// Define a regular expression pattern to match special characters
	pattern := `(\*\-|\+|-|\*|/|\(|\))`
	// Compile the regular expression
	regex, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Println("Error compiling regular expression:", err)
		return input
	}
	// Use the regular expression to find matches and add whitespace around them
	result := regex.ReplaceAllString(input, " $0 ")
	return result
}

//Check if a variable is of any type that can be counted such as int or float
func canBeCounted(v interface{}) bool {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

//Takes a slice and returns the individual elements
func GetInputAsString(values []string) string {
	o := ""
	for _, value := range values {
		// The variable 'value' holds the current element in the slice 'values'
		o += value
	}
	return o
}

var input []string

// Handle the posted form from the template
func InputSave(c *gin.Context) {
	//Recover from panics that may occur during this function's execution
	defer func() {
		if r := recover(); r != nil {
			c.HTML(200, "template.html", gin.H{"Input": "Error"})
			fmt.Println("Recovered from a panic")
		}
	}()

	//Retrieve a session store
	session, _ := store.Get(c.Request, "mysession")

	session.Values["isOperator"] = false
	equals := c.Request.PostFormValue("equals")
	operatorPattern := regexp.MustCompile(`(\*\-|\+|-|\*|\/|\(|\))`)
	foundOperator := false
	operatorPosition := -1
	var periodAfterOperator bool
	var periodInInput bool
	var periodInOperandOne bool

	if equals != "" { //If the equals button has been clicked and a result is being displayed
		session.Values["resultDisplayed"] = true
	}
	for index, item := range input {
		//Check if there is a period in the input box
		if item == "." {
			periodInInput = true
		}
		//Check if there is an operator in the input box and if present get its position
		if operatorPattern.MatchString(item) {
			foundOperator = true
			operatorPosition = index
		}
	}
	//Check if there is a period after an operator in the input box
	if foundOperator {
		for i := operatorPosition + 1; i < len(input); i++ {
			if input[i] == "." {
				periodAfterOperator = true
				break // Exit the loop after finding the first period
			}
		}
	}

	//Iterate over every button clicked and submitted in the html form
	for key, value := range c.Request.PostForm {
		//compile the pattern to used to check for an operator
		pattern := regexp.MustCompile(`(\+|-|\*|\/)`)
		var valueIsOperator bool
		
		//Check if the button pressed is an operator
		for _, valueItem := range value {
			// Check if the pattern matches the current string
			valueIsOperator = pattern.MatchString(valueItem)
		}
		//Check if the button pressed is an operator and the input box was already showing a result
		if (key == "divide" || key == "multiply" || key == "minus" || key == "add") && session.Values["resultDisplayed"] == "yes" {
			session.Values["isOperator"] = true
			session.Values["resultDisplayed"] = "no" // negate this session variable
		}

		switch key {
			case "modulus":
				indexToSlice := operatorPosition + 1
				
				//If the modulus button has been clicked and the input box is empty, display an empty input box 
				if len(input) == 0 {
					input = nil
					c.Redirect(http.StatusSeeOther, "/")
				}
				slicedInput := input[indexToSlice:]                  // Slice the input array just after the operator till the end of the array
				implodeToPercentage := strings.Join(slicedInput, "") // Join slicedinput's elements to a string
				toPercentage, err := strconv.ParseFloat(implodeToPercentage, 64) //Convert the string to float64
				if err != nil {
					// Handle the error
					fmt.Println("Error converting to float:", err)
					input = nil
					input = append(input, "Error")
					c.Redirect(http.StatusSeeOther, "/")
				}

				toPercentage /= 100
				var operandWithoutModulus []string

				if indexToSlice > 0 {
					operandWithoutModulus = input[:indexToSlice] // Slice the input array from beginning to and including the operator
				}
				firstOperatorPattern := regexp.MustCompile(`(\*\-|\+|-|\*|\/|\(|\))`)
				secondOperatorPattern := regexp.MustCompile(`(\*|\/|\(|\))`)
				plusInInput := false
				divisionInInput := false
				negativeInInput := false
				
				//Iterate over the input slice checking the nature of the operators
				for _, element := range input {
					if session.Values["multiAndMinus"] == true {
						negativeInInput = true
					} else if secondOperatorPattern.MatchString(element) {
						divisionInInput = true
					} else if firstOperatorPattern.MatchString(element) {
						plusInInput = true
					}
				}

				if plusInInput { //If the operator is plus or minus symbol
					operandWithoutOperator := input[:operatorPosition] // Slice the input array from beginning to just before the operator
					implodedOperandStr := strings.Join(operandWithoutOperator, "")
					implodedOperand, err := strconv.ParseFloat(implodedOperandStr, 64)
					if err != nil {
						// Handle the error, e.g., log it, return an error value, etc.
						fmt.Println("Error converting to float:", err)
						input = nil
						input = append(input, "Error")
						c.Redirect(http.StatusSeeOther, "/")
					}

					operandWithModulus := implodedOperand * toPercentage
					operandWithoutModulus = append(operandWithoutModulus, strconv.FormatFloat(operandWithModulus, 'f', -1, 64))
					operandWithoutModulusStr := strings.Join(operandWithoutModulus, "")
					currentValueSpc := addWhitespaceAroundOperators(operandWithoutModulusStr)
					currentValue, err := CalculateInput(currentValueSpc)
					if err != nil {
						fmt.Printf("Error evaluating expression: %v\n", err)
					} else {
						fmt.Printf("Result: %.2f\n", currentValue)
					}

					input = nil // Have the input array being empty
					input = append(input, strconv.FormatFloat(currentValue, 'f', -1, 64))
					session.Values["resultDisplayed"] = "yes"
				} else if divisionInInput { //If the operator is a division or multiplication sign
					operandWithoutModulus = append(operandWithoutModulus, strconv.FormatFloat(toPercentage, 'f', -1, 64))
					operandWithoutModulusStr := strings.Join(operandWithoutModulus, "")
					currentValueSpc := addWhitespaceAroundOperators(operandWithoutModulusStr)
					currentValue, err := CalculateInput(currentValueSpc)
					if err != nil {
						fmt.Printf("Error evaluating expression: %v\n", err)
					} else {
						fmt.Printf("Result: %.2f\n", currentValue)
					}
					
					input = nil
					input = append(input, strconv.FormatFloat(currentValue, 'f', -1, 64))
					session.Values["resultDisplayed"] = "yes"
				} else if negativeInInput { //If the operators are a multiplication and a minus sign
					operandWithoutModulus = append(operandWithoutModulus, strconv.FormatFloat(toPercentage, 'f', -1, 64))//Convert to a string before appending to operandWithoutModulus
					operandWithoutModulusStr := strings.Join(operandWithoutModulus, "")
					currentValueSpc := addWhitespaceAroundOperators(operandWithoutModulusStr)
					currentValue, err := CalculateInput(currentValueSpc)
					if err != nil {
						fmt.Printf("Error evaluating expression: %v\n", err)
					} else {
						fmt.Printf("Result: %.2f\n", currentValue) //Display the number with two decimal places
					}

					input = nil //Set the slice to empty
					input = append(input, strconv.FormatFloat(currentValue, 'f', -1, 64))
					session.Values["resultDisplayed"] = "yes"
					session.Values["multiAndMinus"] = false
				} else { // if there is no operator
					currentValueFlt, err := strconv.ParseFloat(GetInputAsString(input), 64)
					if err != nil {
						// Handle the error, e.g., log it, return an error value, etc.
						fmt.Println("Error converting to float:", err)
						return
					}

					currentValueFlt /= 100
					input = nil
					input = append(input, strconv.FormatFloat(currentValueFlt, 'f', -1, 64))
					session.Values["resultDisplayed"] = "yes"
				}

			case "equals":
				inputConv := strings.Join(input, "")
				currentValueSpc := addWhitespaceAroundOperators(inputConv)
				currentValue, err := CalculateInput(currentValueSpc) //Pass along the values for calculation
				if err != nil {
					fmt.Printf("Error evaluating expression: %v\n", err)
				} else {
					fmt.Printf("Result: %.2f\n", currentValue)
				}

				// If the Answer's first value is a period, precede it with a zero
				if strings.HasPrefix(strconv.FormatFloat(currentValue, 'f', -1, 64), ".") {
					currentValueStr := "0" + strconv.FormatFloat(currentValue, 'f', -1, 64)
					currentValueFlt, err := strconv.ParseFloat(currentValueStr, 64)
					if err != nil {
						// Handle the error, e.g., log it, return an error value, etc.
						fmt.Println("Error9 converting to float:", err)
						return
					}
					currentValue = currentValueFlt
				}

				input = nil
				input = append(input, strconv.FormatFloat(currentValue, 'f', -1, 64))
				session.Values["resultDisplayed"] = "yes"

			case "c":
				input = nil //Have the input slice as empty

			case "delete":
				//Remove the last element from the input slice
				if len(input) > 0 {
					input = input[:len(input)-1]
				}

			default:
					var lastElement string
					var lastElementOperator bool
					var secondLastElement int
					var lastElementCountable bool

					if len(input) > 0 {
						lastElement = input[len(input)-1]
						lastElementOperator = operatorPattern.MatchString(lastElement)
						lastElementInt, err := strconv.Atoi(lastElement) //Attempt to convert the last element in the slice to an integer
						if err != nil {
							// Handle the error, for example, by returning a default value or logging the error.
							fmt.Println("Error:", err)
						}
						lastElementCountable = canBeCounted(lastElementInt)
					}
					if len(input) > 2 {
						secondLastElement = len(input) - 2
					}
					secondLastMultiply := false

					if secondLastElement != 0 {
						if input[secondLastElement] == "*" {
							secondLastMultiply = true
						}
					}
					//Check if there is a period in the first operand
					if foundOperator == false && periodInInput && key == "period" {
						periodInOperandOne = true
					}
					//If the period button has been clicked and there is no number preceeding it, append a zero before the period
					if key == "period" && (lastElementCountable == false || len(input) < 1 || lastElementOperator) {
						input = append(input, "0")
					}
					var valueIsMinus bool
					
					//Check if the button pressed is the minus button
					for _, singleValue := range value {
						if singleValue == "-" {
							// Do something when "-" is found in the slice
							valueIsMinus = true
						}
					}
				
					if (key == "divide" || key == "multiply" || key == "add") && valueIsOperator && len(input) < 1 {/*Prohibit an operator
						being provided as the first value in the input box*/
						input = nil
					} else if lastElement == "*" && valueIsMinus { //Allow the minus operator to follow the multiplication operator
						session.Values["multiAndMinus"] = true
						input = append(input, value...) //Append the value slice to the input slice
					} else if secondLastMultiply && lastElement == "-" && valueIsOperator { /*If any operator comes immediately after
						a multiplication sign and a minus sign, remove the latter two and insert that operator into the input box */
						if len(input) >= 2 {
							input = input[:len(input)-2]
						}
						input = append(input, value...)
					} else if valueIsOperator && lastElementOperator { //Prevent two operators from following each other
						if len(input) > 0 {
							input = input[:len(input)-1]
						}
						input = append(input, value...)
					} else if session.Values["isOperator"] == true { /*If an operator is clicked while the input box is displaying an answer
						use that answer for subsequent calculations as the first operand */
						input = append(input, value...)
					} else if session.Values["resultDisplayed"] == "yes" { /*If any number is clicked while the input box is displaying an
						answer, erase and start inserting values afresh */
						input = nil
						input = append(input, value...)
						session.Values["resultDisplayed"] = "no"
					} else if periodInOperandOne || (periodAfterOperator && key == "period") { //Prevent an operand from having more than one period
						input = input
					} else { //Add the value of the clicked button to the input array
						input = append(input, value...)
					}
		}
	}
	session.Save(c.Request, c.Writer) //Save session data 
	c.Redirect(http.StatusSeeOther, "/") //Redirect to the "Display" function via routing
}

//Handle displaying of dynamic data in the html template
func Display(c *gin.Context) {
	session, _ := store.Get(c.Request, "mysession")
	session.Values["inputsession"] = GetInputAsString(input)
	session.Save(c.Request, c.Writer)
	// Go to the html template and send along a map of a key - value pair to be displayed in the html page
	c.HTML(200, "template.html", gin.H{"Input": session.Values["inputsession"]})
}
