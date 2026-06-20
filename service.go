package main

import (
	"errors"
	"fmt"
	"sort"
)

type AddMaterialParams struct {
	ID    string
	Name  string
	Type  MaterialType
	Unit  string
	Stock int
	Cost  int
}

func AddMaterial(store *Store, p AddMaterialParams) error {
	if p.ID == "" {
		return errors.New("材料ID不能为空")
	}
	if p.Name == "" {
		return errors.New("材料名称不能为空")
	}
	if p.Unit == "" {
		return errors.New("单位不能为空")
	}
	if p.Stock < 0 {
		return errors.New("初始库存不能为负数")
	}
	if p.Cost < 0 {
		return errors.New("进货价不能为负数")
	}

	_, err := store.Transaction(func(db *Database) error {
		if existing := FindMaterial(db, p.ID); existing != nil {
			return fmt.Errorf("材料ID %s 已存在", p.ID)
		}
		db.Materials = append(db.Materials, Material{
			ID:    p.ID,
			Name:  p.Name,
			Type:  p.Type,
			Unit:  p.Unit,
			Stock: p.Stock,
			Cost:  p.Cost,
		})
		return nil
	})
	return err
}

type InboundParams struct {
	MaterialID string
	Qty        int
	Cost       int
	Date       string
}

func Inbound(store *Store, p InboundParams) error {
	if p.MaterialID == "" {
		return errors.New("材料ID不能为空")
	}
	if p.Qty <= 0 {
		return errors.New("入库数量必须是正整数")
	}
	if p.Cost <= 0 {
		return errors.New("入库成本必须是正整数")
	}
	if err := ValidateDate(p.Date); err != nil {
		return err
	}

	_, err := store.Transaction(func(db *Database) error {
		m := FindMaterial(db, p.MaterialID)
		if m == nil {
			return fmt.Errorf("材料ID %s 不存在", p.MaterialID)
		}
		m.Stock += p.Qty
		db.InboundRecords = append(db.InboundRecords, InboundRecord{
			MaterialID: p.MaterialID,
			Qty:        p.Qty,
			Cost:       p.Cost,
			Date:       p.Date,
		})
		return nil
	})
	return err
}

type AddOrderParams struct {
	ID         string
	Client     string
	MaterialID string
	Qty        int
	Product    string
	Price      int
	Date       string
}

func AddOrder(store *Store, p AddOrderParams) error {
	if p.ID == "" {
		return errors.New("订单ID不能为空")
	}
	if p.Client == "" {
		return errors.New("客户名称不能为空")
	}
	if p.MaterialID == "" {
		return errors.New("材料ID不能为空")
	}
	if p.Qty <= 0 {
		return errors.New("订单数量必须是正整数")
	}
	if p.Product == "" {
		return errors.New("产品名称不能为空")
	}
	if p.Price <= 0 {
		return errors.New("售价必须是正整数")
	}
	if err := ValidateDate(p.Date); err != nil {
		return err
	}

	_, err := store.Transaction(func(db *Database) error {
		if existing := FindOrder(db, p.ID); existing != nil {
			return fmt.Errorf("订单ID %s 已存在", p.ID)
		}
		m := FindMaterial(db, p.MaterialID)
		if m == nil {
			return fmt.Errorf("材料ID %s 不存在", p.MaterialID)
		}
		if m.Stock < p.Qty {
			return fmt.Errorf("库存不足，当前库存 %d %s，还需 %d %s",
				m.Stock, m.Unit, p.Qty-m.Stock, m.Unit)
		}
		m.Stock -= p.Qty
		db.Orders = append(db.Orders, Order{
			ID:           p.ID,
			Client:       p.Client,
			MaterialID:   p.MaterialID,
			Qty:          p.Qty,
			Product:      p.Product,
			Price:        p.Price,
			Date:         p.Date,
			Status:       StatusPending,
			CompleteDate: "",
		})
		return nil
	})
	return err
}

type CompleteParams struct {
	ID   string
	Date string
}

func CompleteOrder(store *Store, p CompleteParams) error {
	if p.ID == "" {
		return errors.New("订单ID不能为空")
	}
	if err := ValidateDate(p.Date); err != nil {
		return err
	}

	_, err := store.Transaction(func(db *Database) error {
		o := FindOrder(db, p.ID)
		if o == nil {
			return fmt.Errorf("订单ID %s 不存在", p.ID)
		}
		if o.Status != StatusPending {
			return fmt.Errorf("订单当前状态为 %s，只有待加工状态才能完成", o.Status)
		}
		o.Status = StatusFinished
		o.CompleteDate = p.Date
		return nil
	})
	return err
}

type MonthlyResult struct {
	Type       MaterialType
	OrderCount int
	Revenue    int
}

func MonthlyReport(store *Store, month string) ([]MonthlyResult, error) {
	if err := ValidateMonth(month); err != nil {
		return nil, err
	}

	db, err := store.Load()
	if err != nil {
		return nil, err
	}

	matTypeMap := make(map[string]MaterialType)
	for _, m := range db.Materials {
		matTypeMap[m.ID] = m.Type
	}

	resultMap := make(map[MaterialType]*MonthlyResult)
	for _, t := range []MaterialType{TypeNeedleLeaf, TypeBroadLeaf, TypeManMade, TypeBamboo, TypeOther} {
		resultMap[t] = &MonthlyResult{Type: t, OrderCount: 0, Revenue: 0}
	}

	for _, o := range db.Orders {
		if len(o.Date) >= 7 && o.Date[:7] == month {
			mt, ok := matTypeMap[o.MaterialID]
			if !ok {
				continue
			}
			r := resultMap[mt]
			r.OrderCount++
			r.Revenue += o.Price
		}
	}

	results := make([]MonthlyResult, 0, 5)
	for _, t := range []MaterialType{TypeNeedleLeaf, TypeBroadLeaf, TypeManMade, TypeBamboo, TypeOther} {
		results = append(results, *resultMap[t])
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Revenue > results[j].Revenue
	})

	return results, nil
}
