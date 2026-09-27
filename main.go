package main

import "fmt"

func twoSum(nums []int, target int) []int {
	for i, num1 := range nums {
		for j := i + 1; j < len(nums); j++ {
			num2 := nums[j]
			if num1+num2 == target {
				return []int{i, j}
			}
		}
	}
	return nil
}

func main() {
	nums := []int{2, 7, 11, 15}
	target := 9
	result := twoSum(nums, target)
	fmt.Println(result)
}
