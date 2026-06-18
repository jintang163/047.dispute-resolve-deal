package main

import (
	"context"
	"time"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/utils"
	dispute "github.com/dispute-resolve/dispute-service/kitex_gen/dispute"
)

type DisputeServiceImpl struct{}

func (s *DisputeServiceImpl) GetDisputeList(ctx context.Context, req *dispute.GetDisputeListRequest) (resp *dispute.GetDisputeListResponse, err error) {
	resp = &dispute.GetDisputeListResponse{Code: 0, Message: "success"}

	var cases []model.DisputeCase
	var total int64

	db := database.GetDB().Model(&model.DisputeCase{}).Where("deleted_at IS NULL")

	if req.Role == constants.RoleMediator {
		db = db.Where("mediator_id = ?", req.UserId)
	} else if req.Role >= constants.RoleLeader && req.Role <= constants.RoleDirector {
		db = db.Where("org_id = ?", req.OrganizationId)
	}

	if req.Status > 0 {
		db = db.Where("status = ?", req.Status)
	}
	if req.TypeId > 0 {
		db = db.Where("type_id = ?", req.TypeId)
	}
	if req.Keyword != "" {
		db = db.Where("title LIKE ? OR case_no LIKE ? OR applicant_name LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	db.Count(&total)
	offset := int((req.Page - 1) * req.PageSize)
	db.Offset(offset).Limit(int(req.PageSize)).Order("created_at DESC").Find(&cases)

	resp.Total = total
	resp.Cases = make([]*dispute.DisputeCase, len(cases))
	for i, c := range cases {
		resp.Cases[i] = &dispute.DisputeCase{
			Id:              c.ID,
			CaseNo:          c.CaseNo,
			Title:           c.Title,
			TypeId:          c.TypeID,
			CaseLevel:       c.CaseLevel,
			ApplicantName:   c.ApplicantName,
			RespondentName:  c.RespondentName,
			MediatorId:      c.MediatorID,
			MediatorName:    c.MediatorName,
			Status:          c.Status,
			UrgencyLevel:    c.UrgencyLevel,
			CreatedAt:       c.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return resp, nil
}

func (s *DisputeServiceImpl) GetDisputeDetail(ctx context.Context, req *dispute.GetDisputeDetailRequest) (resp *dispute.GetDisputeDetailResponse, err error) {
	resp = &dispute.GetDisputeDetailResponse{Code: 0, Message: "success"}

	var caseData model.DisputeCase
	result := database.GetDB().Where("id = ? AND deleted_at IS NULL", req.CaseId).First(&caseData)
	if result.Error != nil {
		resp.Code = 404
		resp.Message = "案件不存在"
		return resp, nil
	}

	resp.CaseInfo = &dispute.DisputeCase{
		Id:              caseData.ID,
		CaseNo:          caseData.CaseNo,
		Title:           caseData.Title,
		TypeId:          caseData.TypeID,
		CaseLevel:       caseData.CaseLevel,
		Description:     caseData.Description,
		ExpectedSolution: caseData.ExpectedSolution,
		Source:          caseData.Source,
		ApplicantId:     caseData.ApplicantID,
		ApplicantName:   caseData.ApplicantName,
		ApplicantIdcard: caseData.ApplicantIDCard,
		ApplicantMobile: caseData.ApplicantMobile,
		RespondentName:  caseData.RespondentName,
		RespondentIdcard: caseData.RespondentIDCard,
		RespondentMobile: caseData.RespondentMobile,
		Longitude:       caseData.Longitude,
		Latitude:        caseData.Latitude,
		Address:         caseData.Address,
		OrganizationId:  caseData.OrganizationID,
		MediatorId:      caseData.MediatorID,
		MediatorName:    caseData.MediatorName,
		Status:          caseData.Status,
		UrgencyLevel:    caseData.UrgencyLevel,
		CreatedAt:       caseData.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return resp, nil
}

func (s *DisputeServiceImpl) CreateDispute(ctx context.Context, req *dispute.CreateDisputeRequest) (resp *dispute.CreateDisputeResponse, err error) {
	resp = &dispute.CreateDisputeResponse{Code: 0, Message: "success"}

	caseData := &model.DisputeCase{
		CaseNo:          utils.GenerateCaseNo(req.Dispute.OrganizationId),
		Title:           req.Dispute.Title,
		TypeID:          req.Dispute.TypeId,
		Level:           int(req.Dispute.CaseLevel),
		Description:     req.Dispute.Description,
		ExpectedSolution: req.Dispute.ExpectedSolution,
		Source:          req.Dispute.Source,
		ApplicantName:   req.Dispute.ApplicantName,
		ApplicantIDCard: req.Dispute.ApplicantIdcard,
		ApplicantPhone:  req.Dispute.ApplicantMobile,
		RespondentName:  req.Dispute.RespondentName,
		RespondentIDCard: req.Dispute.RespondentIdcard,
		RespondentPhone: req.Dispute.RespondentMobile,
		Longitude:       req.Dispute.Longitude,
		Latitude:        req.Dispute.Latitude,
		Address:         req.Dispute.Address,
		OrgID:           req.Dispute.OrganizationId,
		Status:          constants.CaseStatusPending,
		UrgencyLevel:    int(req.Dispute.UrgencyLevel),
		CreatedBy:       req.Dispute.CreatedBy,
	}

	tx := database.GetDB().Begin()
	if err := tx.Create(caseData).Error; err != nil {
		tx.Rollback()
		resp.Code = 500
		resp.Message = "创建案件失败"
		logger.Error("Create dispute error", logger.Error(err))
		return resp, nil
	}

	if len(req.Evidence) > 0 {
		for _, e := range req.Evidence {
			evidence := &model.DisputeEvidence{
				CaseID:     caseData.ID,
				FileType:   e.FileType,
				FileName:   e.FileName,
				FileURL:    e.FileUrl,
				FileSize:   e.FileSize,
				Remark:     e.Remark,
				SortOrder:  int(e.SortOrder),
				UploadFrom: int(e.UploadFrom),
				UploadedBy: e.UploadedBy,
			}
			if err := tx.Create(evidence).Error; err != nil {
				tx.Rollback()
				resp.Code = 500
				resp.Message = "保存证据失败"
				return resp, nil
			}
		}
	}

	history := &model.DisputeHistory{
		CaseID:     caseData.ID,
		ActionType: constants.HistoryActionCreate,
		ActionName: "创建案件",
		OperatorID: caseData.CreatedBy,
	}
	tx.Create(history)

	tx.Commit()

	resp.CaseId = caseData.ID
	resp.CaseNo = caseData.CaseNo

	return resp, nil
}

func (s *DisputeServiceImpl) AssignDispute(ctx context.Context, req *dispute.AssignDisputeRequest) (resp *dispute.AssignDisputeResponse, err error) {
	resp = &dispute.AssignDisputeResponse{Code: 0, Message: "success"}

	var mediator model.User
	database.GetDB().Select("real_name").Where("id = ?", req.MediatorId).First(&mediator)

	updates := map[string]interface{}{
		"mediator_id":   req.MediatorId,
		"mediator_name": mediator.RealName,
		"status":        constants.CaseStatusMediating,
		"assigned_at":   time.Now(),
	}

	result := database.GetDB().Model(&model.DisputeCase{}).Where("id = ?", req.CaseId).Updates(updates)
	if result.Error != nil {
		resp.Code = 500
		resp.Message = "分派案件失败"
		return resp, nil
	}

	history := &model.DisputeHistory{
		CaseID:     req.CaseId,
		ActionType: constants.HistoryActionAssign,
		ActionName: "分派案件",
		Remark:     "分派给调解员：" + mediator.RealName,
		OperatorID: req.AssignorId,
	}
	database.GetDB().Create(history)

	return resp, nil
}

func (s *DisputeServiceImpl) UrgeDispute(ctx context.Context, req *dispute.UrgeDisputeRequest) (resp *dispute.UrgeDisputeResponse, err error) {
	resp = &dispute.UrgeDisputeResponse{Code: 0, Message: "success"}

	history := &model.DisputeHistory{
		CaseID:     req.CaseId,
		ActionType: constants.HistoryActionUrge,
		ActionName: "催办案件",
		Remark:     req.Remark,
		OperatorID: req.UserId,
	}
	database.GetDB().Create(history)

	return resp, nil
}

func (s *DisputeServiceImpl) UpdateDisputeStatus(ctx context.Context, req *dispute.UpdateDisputeStatusRequest) (resp *dispute.UpdateDisputeStatusResponse, err error) {
	resp = &dispute.UpdateDisputeStatusResponse{Code: 0, Message: "success"}

	updates := map[string]interface{}{
		"status": req.Status,
	}
	if req.Status == constants.CaseStatusClosed {
		updates["closed_at"] = time.Now()
	}

	result := database.GetDB().Model(&model.DisputeCase{}).Where("id = ?", req.CaseId).Updates(updates)
	if result.Error != nil {
		resp.Code = 500
		resp.Message = "更新状态失败"
		return resp, nil
	}

	history := &model.DisputeHistory{
		CaseID:     req.CaseId,
		ActionType: constants.HistoryActionStatus,
		ActionName: "更新状态",
		Remark:     req.Remark,
		OperatorID: req.UserId,
	}
	database.GetDB().Create(history)

	return resp, nil
}

func (s *DisputeServiceImpl) GetDisputeHistory(ctx context.Context, req *dispute.GetDisputeHistoryRequest) (resp *dispute.GetDisputeHistoryResponse, err error) {
	resp = &dispute.GetDisputeHistoryResponse{Code: 0, Message: "success"}

	var history []model.DisputeHistory
	database.GetDB().Where("case_id = ?", req.CaseId).Order("created_at ASC").Find(&history)

	resp.History = make([]*dispute.DisputeHistory, len(history))
	for i, h := range history {
		resp.History[i] = &dispute.DisputeHistory{
			Id:           h.ID,
			CaseId:       h.CaseID,
			ActionType:   h.ActionType,
			ActionName:   h.ActionName,
			Remark:       h.Remark,
			OperatorId:   h.OperatorID,
			OperatorName: h.OperatorName,
			CreatedAt:    h.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return resp, nil
}

func (s *DisputeServiceImpl) GetDisputeProgress(ctx context.Context, req *dispute.GetDisputeProgressRequest) (resp *dispute.GetDisputeProgressResponse, err error) {
	resp = &dispute.GetDisputeProgressResponse{Code: 0, Message: "success"}

	var caseData model.DisputeCase
	result := database.GetDB().Where("case_no = ? AND applicant_idcard = ? AND deleted_at IS NULL",
		req.CaseNo, req.Idcard).First(&caseData)
	if result.Error != nil {
		resp.Code = 404
		resp.Message = "案件不存在或信息不匹配"
		return resp, nil
	}

	resp.CaseInfo = &dispute.DisputeCase{
		Id:        caseData.ID,
		CaseNo:    caseData.CaseNo,
		Title:     caseData.Title,
		Status:    caseData.Status,
		CreatedAt: caseData.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	var history []model.DisputeHistory
	database.GetDB().Where("case_id = ?", caseData.ID).Order("created_at ASC").Find(&history)

	resp.History = make([]*dispute.DisputeHistory, len(history))
	for i, h := range history {
		resp.History[i] = &dispute.DisputeHistory{
			Id:           h.ID,
			ActionType:   h.ActionType,
			ActionName:   h.ActionName,
			Remark:       h.Remark,
			OperatorName: h.OperatorName,
			CreatedAt:    h.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return resp, nil
}

func (s *DisputeServiceImpl) GetDisputeTypes(ctx context.Context) (resp *dispute.GetDisputeTypesResponse, err error) {
	resp = &dispute.GetDisputeTypesResponse{Code: 0, Message: "success"}

	var types []model.DisputeType
	database.GetDB().Where("status = 1 AND deleted_at IS NULL").Order("level, sort_order").Find(&types)

	resp.Types = make([]*dispute.DisputeType, len(types))
	for i, t := range types {
		resp.Types[i] = &dispute.DisputeType{
			Id:          t.ID,
			Name:        t.Name,
			Code:        t.Code,
			ParentId:    t.ParentID,
			Level:       t.Level,
			Description: t.Description,
			SortOrder:   t.SortOrder,
			Status:      t.Status,
		}
	}

	return resp, nil
}

func (s *DisputeServiceImpl) KioskCreateDispute(ctx context.Context, req *dispute.KioskCreateDisputeRequest) (resp *dispute.CreateDisputeResponse, err error) {
	createReq := &dispute.CreateDisputeRequest{
		Dispute:  req.Dispute,
		Evidence: req.Evidence,
	}
	return s.CreateDispute(ctx, createReq)
}

func (s *DisputeServiceImpl) MiniAppCreateDispute(ctx context.Context, req *dispute.CreateDisputeRequest) (resp *dispute.CreateDisputeResponse, err error) {
	return s.CreateDispute(ctx, req)
}

func (s *DisputeServiceImpl) UploadEvidence(ctx context.Context, req *dispute.UploadEvidenceRequest) (resp *dispute.UploadEvidenceResponse, err error) {
	resp = &dispute.UploadEvidenceResponse{Code: 0, Message: "success"}

	evidence := &model.DisputeEvidence{
		CaseID:     req.CaseId,
		FileType:   req.FileType,
		FileName:   req.FileName,
		FileURL:    req.FileUrl,
		FileSize:   req.FileSize,
		Remark:     req.Remark,
		SortOrder:  0,
		UploadFrom: int(req.UploadFrom),
		UploadedBy: req.UserId,
	}

	result := database.GetDB().Create(evidence)
	if result.Error != nil {
		resp.Code = 500
		resp.Message = "上传证据失败"
		return resp, nil
	}

	resp.Evidence = &dispute.Evidence{
		Id:         evidence.ID,
		CaseId:     evidence.CaseID,
		FileType:   evidence.FileType,
		FileName:   evidence.FileName,
		FileUrl:    evidence.FileURL,
		FileSize:   evidence.FileSize,
		Remark:     evidence.Remark,
		CreatedAt:  evidence.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return resp, nil
}

func (s *DisputeServiceImpl) GetEvidenceList(ctx context.Context, req *dispute.GetEvidenceListRequest) (resp *dispute.GetEvidenceListResponse, err error) {
	resp = &dispute.GetEvidenceListResponse{Code: 0, Message: "success"}

	var evidence []model.Evidence
	var total int64

	db := database.GetDB().Model(&model.Evidence{}).Where("case_id = ? AND deleted_at IS NULL", req.CaseId)
	db.Count(&total)

	offset := int((req.Page - 1) * req.PageSize)
	db.Offset(offset).Limit(int(req.PageSize)).Order("sort_order, created_at DESC").Find(&evidence)

	resp.Total = total
	resp.Evidence = make([]*dispute.Evidence, len(evidence))
	for i, e := range evidence {
		resp.Evidence[i] = &dispute.Evidence{
			Id:         e.ID,
			CaseId:     e.CaseID,
			FileType:   e.FileType,
			FileName:   e.FileName,
			FileUrl:    e.FileURL,
			FileSize:   e.FileSize,
			Remark:     e.Remark,
			SortOrder:  int32(e.SortOrder),
			UploadFrom: int32(e.UploadFrom),
			CreatedAt:  e.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return resp, nil
}

func (s *DisputeServiceImpl) DeleteEvidence(ctx context.Context, req *dispute.DeleteEvidenceRequest) (resp *dispute.DeleteEvidenceResponse, err error) {
	resp = &dispute.DeleteEvidenceResponse{Code: 0, Message: "success"}

	result := database.GetDB().Model(&model.DisputeEvidence{}).Where("id = ?", req.EvidenceId).Update("deleted_at", time.Now())
	if result.Error != nil {
		resp.Code = 500
		resp.Message = "删除证据失败"
		return resp, nil
	}

	return resp, nil
}

func (s *DisputeServiceImpl) BatchDeleteEvidence(ctx context.Context, req *dispute.BatchDeleteEvidenceRequest) (resp *dispute.BatchDeleteEvidenceResponse, err error) {
	resp = &dispute.BatchDeleteEvidenceResponse{Code: 0, Message: "success"}

	result := database.GetDB().Model(&model.Evidence{}).Where("id IN ?", req.EvidenceIds).Update("deleted_at", time.Now())
	if result.Error != nil {
		resp.Code = 500
		resp.Message = "批量删除证据失败"
		return resp, nil
	}

	resp.DeletedCount = result.RowsAffected
	return resp, nil
}

func (s *DisputeServiceImpl) UpdateEvidenceRemark(ctx context.Context, req *dispute.UpdateEvidenceRemarkRequest) (resp *dispute.UpdateEvidenceRemarkResponse, err error) {
	resp = &dispute.UpdateEvidenceRemarkResponse{Code: 0, Message: "success"}

	result := database.GetDB().Model(&model.DisputeEvidence{}).Where("id = ?", req.EvidenceId).Update("remark", req.Remark)
	if result.Error != nil {
		resp.Code = 500
		resp.Message = "更新备注失败"
		return resp, nil
	}

	return resp, nil
}
