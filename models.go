package main

import (
	"errors"
	"strings"
)

type MaterialType string

const (
	TypeNeedleLeaf  MaterialType = "针叶木"
	TypeBroadLeaf   MaterialType = "阔叶木"
	TypeManMade     MaterialType = "人造板"
	TypeBamboo      MaterialType = "竹材"
	TypeOther       MaterialType = "其他"
)

var ValidMaterialTypes = map[MaterialType]bool{
	TypeNeedleLeaf: true,
	TypeBroadLeaf:  true,
	TypeManMade:    true,
	TypeBamboo:     true,
	TypeOther:      true,
}

func (mt MaterialType) IsValid() bool {
	return ValidMaterialTypes[mt]
}

func ParseMaterialType(s string) (MaterialType, error) {
	t := MaterialType(strings.TrimSpace(s))
	if t.IsValid() {
		return t, nil
	}
	return "", errors.New("材料类型只能是：针叶木、阔叶木、人造板、竹材、其他")
}

func AllMaterialTypes() []MaterialType {
	return []MaterialType{TypeNeedleLeaf, TypeBroadLeaf, TypeManMade, TypeBamboo, TypeOther}
}

type OrderStatus string

const (
	StatusPending  OrderStatus = "待加工"
	StatusFinished OrderStatus = "已完成"
)

type Material struct {
	ID    string       `json:"id"`
	Name  string       `json:"name"`
	Type  MaterialType `json:"type"`
	Unit  string       `json:"unit"`
	Stock int          `json:"stock"`
	Cost  int          `json:"cost"`
}

type InboundRecord struct {
	MaterialID string `json:"material_id"`
	Qty        int    `json:"qty"`
	Cost       int    `json:"cost"`
	Date       string `json:"date"`
}

type Order struct {
	ID           string      `json:"id"`
	Client       string      `json:"client"`
	MaterialID   string      `json:"material_id"`
	Qty          int         `json:"qty"`
	Product      string      `json:"product"`
	Price        int         `json:"price"`
	Date         string      `json:"date"`
	Status       OrderStatus `json:"status"`
	CompleteDate string      `json:"complete_date,omitempty"`
}

type Database struct {
	Materials       []Material       `json:"materials"`
	InboundRecords  []InboundRecord  `json:"inbound_records"`
	Orders          []Order          `json:"orders"`
}
