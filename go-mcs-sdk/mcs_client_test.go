package go_mcs_sdk

import (
	"fmt"
	"math/big"
	"testing"
)

func TestApprove(t *testing.T) {

	s := NewMcsClient("https://rpc-mumbai.maticvigil.com")
	err := s.SetAccount("")
	if err != nil {
		t.Errorf("failed to set account: %s", err.Error())
		return
	}
	allowance := s.queryAllowance()
	if allowance.Cmp(big.NewInt(0)) <= 0 {
		amount, _ := new(big.Int).SetString("152d02c7e14af6000000", 16)
		_, err := s.approve(amount)
		if err != nil {
			t.Errorf("failed to approve : %s", err.Error())
		}
	}
}

func TestLock(t *testing.T) {

	s := NewMcsClient("https://rpc-mumbai.maticvigil.com")
	err := s.SetAccount("")
	if err != nil {
		t.Errorf("failed to set account: %s", err.Error())
		return
	}
	allowance := s.queryAllowance()
	amount := big.NewInt(9554660000000000)
	if allowance.Cmp(amount) <= 0 {
		_, err := s.approve(new(big.Int).Mul(amount, big.NewInt(100000000)))
		if err != nil {
			t.Errorf("failed to approve : %s", err.Error())
		}
	}
	tx, err := s.lockToken("", amount, 34311)
	fmt.Println(tx, err)
}

func TestPayment(t *testing.T) {

	s := NewMscClient("https://rpc-mumbai.maticvigil.com")
	err := s.SetAccount("")
	if err != nil {
		t.Errorf("failed to set account: %s", err.Error())
		return
	}
	tx, err := s.MakePayment("", 34311, 540)
	fmt.Println(tx, err)
}
