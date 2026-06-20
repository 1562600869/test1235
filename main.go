package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

func printUsage() {
	fmt.Println(`社区木材加工坊管理工具

用法:
  go run main.go <命令> [参数]

命令:
  add-material  添加原材料
  inbound       原材料入库
  add-order     接单（扣减库存）
  complete      完成订单
  monthly       月度统计报表

示例:
  go run main.go add-material M001 松木板 --type 针叶木 --unit 张 --stock 80 --cost 35
  go run main.go inbound M001 --qty 40 --cost 35 --date 2024-03-20
  go run main.go add-order O001 --client 张先生 --material M001 --qty 10 --product 书架 --price 280 --date 2024-03-20
  go run main.go complete O001 --date 2024-03-25
  go run main.go monthly --month 2024-03
`)
}

func parseAddMaterial(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("用法: add-material <ID> <名称> --type 类型 --unit 单位 --stock 库存 --cost 进货价")
	}
	id := args[0]
	name := args[1]

	fs := flag.NewFlagSet("add-material", flag.ExitOnError)
	typeStr := fs.String("type", "", "材料类型：针叶木/阔叶木/人造板/竹材/其他")
	unit := fs.String("unit", "", "单位，如：张、立方米、根等")
	stockStr := fs.String("stock", "0", "初始库存数量，正整数")
	costStr := fs.String("cost", "0", "进货价，整数分/单位")

	if err := fs.Parse(args[2:]); err != nil {
		return err
	}

	mt, err := ParseMaterialType(*typeStr)
	if err != nil {
		return err
	}

	stock, err := strconv.Atoi(*stockStr)
	if err != nil {
		return fmt.Errorf("库存数量必须是整数: %v", err)
	}
	cost, err := strconv.Atoi(*costStr)
	if err != nil {
		return fmt.Errorf("进货价必须是整数: %v", err)
	}

	params := AddMaterialParams{
		ID:    id,
		Name:  name,
		Type:  mt,
		Unit:  *unit,
		Stock: stock,
		Cost:  cost,
	}
	if err := AddMaterial(NewStore(), params); err != nil {
		return err
	}

	fmt.Printf("材料添加成功: [%s] %s (%s)\n", id, name, mt)
	return nil
}

func parseInbound(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("用法: inbound <材料ID> --qty 数量 --cost 成本 --date YYYY-MM-DD")
	}
	materialID := args[0]

	fs := flag.NewFlagSet("inbound", flag.ExitOnError)
	qtyStr := fs.String("qty", "", "入库数量，正整数")
	costStr := fs.String("cost", "", "入库成本，正整数分")
	date := fs.String("date", "", "入库日期 YYYY-MM-DD")

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	if *qtyStr == "" {
		return fmt.Errorf("--qty 参数必填")
	}
	if *costStr == "" {
		return fmt.Errorf("--cost 参数必填")
	}
	if *date == "" {
		return fmt.Errorf("--date 参数必填")
	}

	qty, err := strconv.Atoi(*qtyStr)
	if err != nil {
		return fmt.Errorf("数量必须是正整数: %v", err)
	}
	cost, err := strconv.Atoi(*costStr)
	if err != nil {
		return fmt.Errorf("成本必须是正整数: %v", err)
	}

	params := InboundParams{
		MaterialID: materialID,
		Qty:        qty,
		Cost:       cost,
		Date:       *date,
	}
	if err := Inbound(NewStore(), params); err != nil {
		return err
	}

	fmt.Printf("入库成功: 材料 %s 入库 %d, 成本 %d 分, 日期 %s\n", materialID, qty, cost, *date)
	return nil
}

func parseAddOrder(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("用法: add-order <订单ID> --client 客户 --material 材料ID --qty 数量 --product 产品 --price 售价 --date YYYY-MM-DD")
	}
	orderID := args[0]

	fs := flag.NewFlagSet("add-order", flag.ExitOnError)
	client := fs.String("client", "", "客户名称")
	materialID := fs.String("material", "", "材料ID")
	qtyStr := fs.String("qty", "", "订单数量，正整数")
	product := fs.String("product", "", "产品名称")
	priceStr := fs.String("price", "", "售价，整数分")
	date := fs.String("date", "", "接单日期 YYYY-MM-DD")

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	if *client == "" {
		return fmt.Errorf("--client 参数必填")
	}
	if *materialID == "" {
		return fmt.Errorf("--material 参数必填")
	}
	if *qtyStr == "" {
		return fmt.Errorf("--qty 参数必填")
	}
	if *product == "" {
		return fmt.Errorf("--product 参数必填")
	}
	if *priceStr == "" {
		return fmt.Errorf("--price 参数必填")
	}
	if *date == "" {
		return fmt.Errorf("--date 参数必填")
	}

	qty, err := strconv.Atoi(*qtyStr)
	if err != nil {
		return fmt.Errorf("数量必须是正整数: %v", err)
	}
	price, err := strconv.Atoi(*priceStr)
	if err != nil {
		return fmt.Errorf("售价必须是正整数: %v", err)
	}

	params := AddOrderParams{
		ID:         orderID,
		Client:     *client,
		MaterialID: *materialID,
		Qty:        qty,
		Product:    *product,
		Price:      price,
		Date:       *date,
	}
	if err := AddOrder(NewStore(), params); err != nil {
		return err
	}

	fmt.Printf("接单成功: 订单 %s 客户 %s 产品 %s 数量 %d 售价 %d 分\n", orderID, *client, *product, qty, price)
	return nil
}

func parseComplete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("用法: complete <订单ID> --date YYYY-MM-DD")
	}
	orderID := args[0]

	fs := flag.NewFlagSet("complete", flag.ExitOnError)
	date := fs.String("date", "", "完成日期 YYYY-MM-DD")

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	if *date == "" {
		return fmt.Errorf("--date 参数必填")
	}

	params := CompleteParams{
		ID:   orderID,
		Date: *date,
	}
	if err := CompleteOrder(NewStore(), params); err != nil {
		return err
	}

	fmt.Printf("订单完成: %s, 完成日期 %s\n", orderID, *date)
	return nil
}

func parseMonthly(args []string) error {
	fs := flag.NewFlagSet("monthly", flag.ExitOnError)
	month := fs.String("month", "", "月份 YYYY-MM")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *month == "" {
		return fmt.Errorf("--month 参数必填，格式 YYYY-MM")
	}

	results, err := MonthlyReport(NewStore(), *month)
	if err != nil {
		return err
	}

	fmt.Printf("===== %s 月度统计 =====\n", *month)
	fmt.Printf("%-8s  %-10s  %-15s\n", "材料类型", "订单数", "总收入(分)")
	fmt.Println("----------------------------------------")
	totalOrders := 0
	totalRevenue := 0
	for _, r := range results {
		fmt.Printf("%-8s  %-10d  %-15d\n", r.Type, r.OrderCount, r.Revenue)
		totalOrders += r.OrderCount
		totalRevenue += r.Revenue
	}
	fmt.Println("----------------------------------------")
	fmt.Printf("%-8s  %-10d  %-15d\n", "合计", totalOrders, totalRevenue)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "add-material":
		err = parseAddMaterial(args)
	case "inbound":
		err = parseInbound(args)
	case "add-order":
		err = parseAddOrder(args)
	case "complete":
		err = parseComplete(args)
	case "monthly":
		err = parseMonthly(args)
	case "-h", "--help", "help":
		printUsage()
		return
	default:
		fmt.Printf("未知命令: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
