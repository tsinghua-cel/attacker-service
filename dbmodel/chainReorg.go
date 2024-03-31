package dbmodel

import (
	"github.com/astaxie/beego/orm"
	"github.com/tsinghua-cel/attacker-service/types"
	"strconv"
)

type ChainReorg struct {
	ID           int64  `orm:"column(id)" db:"id" json:"id" form:"id"`                                                 //  任务类型id
	Epoch        int64  `orm:"column(epoch)" db:"epoch" json:"epoch" form:"epoch"`                                     // epoch
	Slot         int64  `orm:"column(slot)" db:"slot" json:"slot" form:"slot"`                                         // slot
	Depth        int    `orm:"column(depth)" db:"depth" json:"depth" form:"depth"`                                     // depth
	OldHeadBlock string `orm:"column(old_head_block)" db:"old_head_block" json:"old_head_block" form:"old_head_block"` // old_head_block
	NewHeadBlock string `orm:"column(new_head_block)" db:"new_head_block" json:"new_head_block" form:"new_head_block"` // new_head_block
	OldHeadState string `orm:"column(old_head_state)" db:"old_head_state" json:"old_head_state" form:"old_head_state"` // old_head_state
	NewHeadState string `orm:"column(new_head_state)" db:"new_head_state" json:"new_head_state" form:"new_head_state"` // new_head_state
}

func (ChainReorg) TableName() string {
	return "t_chain_reorg"
}

type ChainReorgRepository interface {
	Create(reorg *ChainReorg) error
	GetListByFilter(filters ...interface{}) []*ChainReorg
}

type chainReorgRepositoryImpl struct {
	o orm.Ormer
}

func NewChainReorgRepository(o orm.Ormer) ChainReorgRepository {
	return &chainReorgRepositoryImpl{o}
}

func (repo *chainReorgRepositoryImpl) Create(reorg *ChainReorg) error {
	_, err := repo.o.Insert(reorg)
	return err
}

func (repo *chainReorgRepositoryImpl) GetListByFilter(filters ...interface{}) []*ChainReorg {
	list := make([]*ChainReorg, 0)
	query := repo.o.QueryTable(new(ChainReorg).TableName())
	if len(filters) > 0 {
		l := len(filters)
		for k := 0; k < l; k += 2 {
			query = query.Filter(filters[k].(string), filters[k+1])
		}
	}
	query.OrderBy("-slot").All(&list)
	return list
}

func InsertNewReorg(reorg types.ReorgEvent) {
	epoch, _ := strconv.ParseInt(reorg.Epoch, 10, 64)
	slot, _ := strconv.ParseInt(reorg.Slot, 10, 64)
	depth, _ := strconv.Atoi(reorg.Depth)
	NewChainReorgRepository(orm.NewOrm()).Create(&ChainReorg{
		Epoch:        epoch,
		Slot:         slot,
		Depth:        depth,
		OldHeadBlock: reorg.OldHeadBlock,
		NewHeadBlock: reorg.NewHeadBlock,
		OldHeadState: reorg.OldHeadState,
		NewHeadState: reorg.NewHeadState,
	})
}

func GetAllReorgList() []*ChainReorg {
	return NewChainReorgRepository(orm.NewOrm()).GetListByFilter()
}
