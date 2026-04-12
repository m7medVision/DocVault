package search

import "fmt"

type FilterBuilder struct {
	conditions []string
	args       []interface{}
	paramCount int
}

func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{
		conditions: []string{},
		args:       []interface{}{},
		paramCount: 0,
	}
}

func (b *FilterBuilder) AddCondition(cond string, args ...interface{}) *FilterBuilder {
	b.conditions = append(b.conditions, cond)
	b.args = append(b.args, args...)
	return b
}

func (b *FilterBuilder) addParam(param interface{}) string {
	b.paramCount++
	b.args = append(b.args, param)
	return fmt.Sprintf("$%d", b.paramCount+1)
}

func (b *FilterBuilder) Build() (string, []interface{}) {
	if len(b.conditions) == 0 {
		return "", nil
	}
	result := b.conditions[0]
	for i := 1; i < len(b.conditions); i++ {
		result += " AND " + b.conditions[i]
	}
	return result, b.args
}
