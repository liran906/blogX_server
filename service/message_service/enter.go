// Path: ./service/message_service/enter.go

package message_service

import (
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum/notify_enum"
	"blogX_server/utils"
	"fmt"
)

// 进来内部函数在再 preload  cmt的fk，总是会报错
// 怀疑是 mysql 没有那么快写入并可读取
// 所以在外部把相关字段填好再传入吧
// 这里不再 Preload 了

func SendCommentNotify(cmt models.CommentModel) (err error) {
	// 判断消息种类
	var messageType notify_enum.Type
	var receiveUserID uint
	if cmt.ParentID == nil {
		messageType = notify_enum.ArticleCommentType
		receiveUserID = cmt.ArticleModel.UserID
	} else {
		messageType = notify_enum.CommentReplyType
		receiveUserID = cmt.ParentModel.UserID
	}

	// 自己评论自己，就不通知了
	if cmt.UserID == receiveUserID {
		return
	}

	// 检验对方是否接受消息
	var receiveUserConf models.UserMessageConfModel
	err = global.DB.Take(&receiveUserConf, "user_id = ?", receiveUserID).Error
	if !receiveUserConf.ReceiveCommentNotify {
		return
	}

	// 加载发送方信息
	var user models.UserModel
	err = global.DB.Where("id = ?", cmt.UserID).Take(&user).Error
	if err != nil {
		return
	}
	cmt.UserModel = user

	// 入库
	err = global.DB.Create(&models.NotifyModel{
		Type:                messageType,
		Content:             utils.ExtractContent(cmt.Content, 30), // 限制最长字数
		ReceiveUserID:       receiveUserID,
		ActionUserID:        cmt.UserID,
		ActionUserNickname:  cmt.UserModel.Nickname,
		ActionUserAvatarURL: cmt.UserModel.AvatarURL,
		ArticleID:           cmt.ArticleID,
		ArticleTitle:        cmt.ArticleModel.Title,
		CommentID:           cmt.ID,
	}).Error
	return
}

func SendArticleLikeNotify(al models.ArticleLikesModel) (err error) {
	// 自己赞自己，就不通知了
	if al.UserID == al.ArticleModel.UserID {
		return
	}

	// 检验对方是否接受消息
	var receiveUserConf models.UserMessageConfModel
	err = global.DB.Take(&receiveUserConf, "user_id = ?", al.ArticleModel.UserID).Error
	if !receiveUserConf.ReceiveLikeNotify {
		return
	}

	// 同个人给同一篇文章点过赞了，就不新发消息了
	err = global.DB.Take(&models.NotifyModel{}, "type = ? AND article_id = ? AND action_user_id = ?", notify_enum.ArticleLikeType, al.ArticleID, al.UserID).Error
	if err == nil {
		return
	}

	// 加载发送方信息
	var user models.UserModel
	err = global.DB.Where("id = ?", al.UserID).Take(&user).Error
	if err != nil {
		return
	}
	al.UserModel = user

	// 入库
	err = global.DB.Create(&models.NotifyModel{
		Type:                notify_enum.ArticleLikeType,
		ReceiveUserID:       al.ArticleModel.UserID,
		ActionUserID:        al.UserID,
		ActionUserNickname:  al.UserModel.Nickname,
		ActionUserAvatarURL: al.UserModel.AvatarURL,
		ArticleID:           al.ArticleID,
		ArticleTitle:        al.ArticleModel.Title,
	}).Error
	return
}

func SendArticleCollectNotify(ac models.ArticleCollectionModel) (err error) {
	// 自己收藏自己，就不通知了
	if ac.UserID == ac.ArticleModel.UserID {
		return
	}

	// 检验对方是否接受消息
	var receiveUserConf models.UserMessageConfModel
	err = global.DB.Take(&receiveUserConf, "user_id = ?", ac.ArticleModel.UserID).Error
	if !receiveUserConf.ReceiveCollectNotify {
		return
	}

	// 同个人给同一篇文章收藏过，就不新发消息了
	err = global.DB.Take(&models.NotifyModel{}, "type = ? AND article_id = ? AND action_user_id = ?", notify_enum.ArticleCollectType, ac.ArticleID, ac.UserID).Error
	if err == nil {
		return
	}

	// 加载发送方信息
	var user models.UserModel
	err = global.DB.Where("id = ?", ac.UserID).Take(&user).Error
	if err != nil {
		return
	}
	ac.UserModel = user

	// 入库
	err = global.DB.Create(&models.NotifyModel{
		Type:                notify_enum.ArticleCollectType,
		ReceiveUserID:       ac.ArticleModel.UserID,
		ActionUserID:        ac.UserID,
		ActionUserNickname:  ac.UserModel.Nickname,
		ActionUserAvatarURL: ac.UserModel.AvatarURL,
		ArticleID:           ac.ArticleID,
		ArticleTitle:        ac.ArticleModel.Title,
	}).Error
	return
}

func SendCommentLikeNotify(cl models.CommentLikesModel) (err error) {
	// 自己赞自己，就不通知了
	if cl.UserID == cl.CommentModel.UserID {
		return
	}

	// 检验对方是否接受消息
	var receiveUserConf models.UserMessageConfModel
	err = global.DB.Take(&receiveUserConf, "user_id = ?", cl.CommentModel.UserID).Error
	if !receiveUserConf.ReceiveLikeNotify {
		return
	}

	// 同个人给同一篇评论点过赞了，就不新发消息了
	err = global.DB.Take(&models.NotifyModel{}, "type = ? AND comment_id = ? AND action_user_id = ?", notify_enum.CommentLikeType, cl.CommentID, cl.UserID).Error
	if err == nil {
		return
	}

	// 加载发送方信息
	var user models.UserModel
	err = global.DB.Where("id = ?", cl.UserID).Take(&user).Error
	if err != nil {
		return
	}
	cl.UserModel = user

	// 入库
	err = global.DB.Create(&models.NotifyModel{
		Type:                notify_enum.CommentLikeType,
		ReceiveUserID:       cl.CommentModel.UserID,
		ActionUserID:        cl.UserID,
		ActionUserNickname:  cl.UserModel.Nickname,
		ActionUserAvatarURL: cl.UserModel.AvatarURL,
		CommentID:           cl.CommentID,
		CommentContent:      utils.ExtractContent(cl.CommentModel.Content, 30),
	}).Error
	return
}

func SendSystemNotify(receiver uint, title, content, link, href string) error {
	// todo 被删除的文章 评论的 id 要记录下
	var user models.UserModel
	err := global.DB.Where("id = ?", receiver).Take(&user).Error
	if err != nil {
		return fmt.Errorf("用户不存在: %s", err)
	}

	var msg = models.NotifyModel{
		Type:          notify_enum.SystemType,
		ReceiveUserID: receiver,
		Title:         title,
		Content:       content,
		LinkLabel:     link,
		LinkHref:      href,
	}
	err = global.DB.Create(&msg).Error
	if err != nil {
		return fmt.Errorf("写入数据库失败: %s", err)
	}
	return nil
}
