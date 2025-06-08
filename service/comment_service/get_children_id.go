// Path: ./service/comment_service/get_children_id.go

package comment_service

import (
	"blogX_server/global"
	"blogX_server/models"
	"errors"
)

// 返回 ID 切片

// GetChildrenCidByID 查询当前节点的子节点，直接数据库搜索
func GetChildrenCidByID(commentID *uint) (childrenCid []uint, err error) {
	if commentID == nil {
		return nil, errors.New("imported comment id is nil")
	}
	err = global.DB.Model(&models.CommentModel{}).Where("parent_id = ?", commentID).Pluck("id", &childrenCid).Error
	if err != nil {
		return
	}
	return
}

// GetOffspringCidByModel 查询当前节点的子节点，数据库先 preload，再搜索
func GetOffspringCidByModel(c *models.CommentModel) (childrenCid []uint, err error) {
	if c == nil {
		return nil, errors.New("imported model id is nil")
	}
	err = global.DB.Model(c).Association("ChildListModel").Find(&c.ChildListModel)
	if err != nil {
		return
	}
	for _, child := range c.ChildListModel {
		childrenCid = append(childrenCid, child.ID)
	}
	return
}

// GetOffspringCidOfRoot 查询根节点的所有子孙节点，传入必须是根节点 id，然后直接数据库查询
func GetOffspringCidOfRoot(rootID *uint) (childrenCid []uint, err error) {
	if rootID == nil {
		return nil, errors.New("imported root id is nil")
	}
	err = global.DB.Model(&models.CommentModel{}).Where("root_id = ?", rootID).Pluck("id", &childrenCid).Error
	if err != nil {
		return
	}
	return
}

// GetOffspringCid 递归查询所有当前节点的所有子孙节点
func GetOffspringCid(commentID *uint) (childrenCid []uint, err error) {
	if commentID == nil {
		return
	}
	var cmt models.CommentModel
	err = global.DB.Preload("ChildListModel").Take(&cmt, commentID).Error
	if err != nil {
		return
	}
	for _, child := range cmt.ChildListModel {
		// 先加入当前节点的子节点
		childrenCid = append(childrenCid, child.ID)

		// 递归查询
		cids, err := GetOffspringCid(&child.ID)
		if err != nil {
			return nil, err
		}
		// 加入递归结果
		childrenCid = append(childrenCid, cids...)
	}
	return
}

// 返回 model 切片

// GetChildrenByID 查询当前节点的子节点，直接数据库搜索
func GetChildrenByID(commentID *uint) (children []models.CommentModel, err error) {
	if commentID == nil {
		return nil, errors.New("imported comment id is nil")
	}
	err = global.DB.Model(&models.CommentModel{}).Where("parent_id = ?", commentID).Find(&children).Error
	if err != nil {
		return
	}
	return
}

// GetOffspringByModel 查询当前节点的子节点，数据库先 preload，再搜索
func GetOffspringByModel(c *models.CommentModel) (offsprings []models.CommentModel, err error) {
	if c == nil {
		return nil, errors.New("imported model id is nil")
	}
	err = global.DB.Model(c).Association("ChildListModel").Find(&c.ChildListModel)
	if err != nil {
		return
	}
	for _, child := range c.ChildListModel {
		offsprings = append(offsprings, *child)
	}
	return
}

// GetOffspringOfRoot 查询根节点的所有子孙节点，传入必须是根节点 id，然后直接数据库查询
func GetOffspringOfRoot(rootID *uint) (offsprings []models.CommentModel, err error) {
	if rootID == nil {
		return nil, errors.New("imported root id is nil")
	}
	err = global.DB.Model(&models.CommentModel{}).Where("root_id = ?", rootID).Find(&offsprings).Error
	if err != nil {
		return
	}
	return
}

// GetOffsprings 递归查询所有当前节点的所有子孙节点
func GetOffsprings(commentID *uint) (offsprings []models.CommentModel, err error) {
	if commentID == nil {
		return
	}
	var cmt models.CommentModel
	err = global.DB.Preload("ChildListModel").Take(&cmt, commentID).Error
	if err != nil {
		return
	}
	for _, child := range cmt.ChildListModel {
		// 先加入当前节点的子节点
		offsprings = append(offsprings, *child)

		// 递归查询
		children, err := GetOffsprings(&child.ID)
		if err != nil {
			return nil, err
		}
		// 加入递归结果
		offsprings = append(offsprings, children...)
	}
	return
}
