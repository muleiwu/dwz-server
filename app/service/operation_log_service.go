package service

import (
	"encoding/json"
	"time"

	"cnb.cool/mliev/open/dwz-server/app/dao"
	"cnb.cool/mliev/open/dwz-server/app/dto"
	"cnb.cool/mliev/open/dwz-server/app/model"
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

type OperationLogService struct {
	logDAO *dao.OperationLogDAO
}

func NewOperationLogService(helper interfaces.HelperInterface) *OperationLogService {
	return &OperationLogService{
		logDAO: dao.NewOperationLogDAO(helper),
	}
}

// CreateLog 创建操作日志
func (s *OperationLogService) CreateLog(userID *uint64, username, operation, resource, resourceID, method, path, requestBody, responseBody, ip, userAgent string, responseCode int, executeTime int64, status int8, errorMessage string) error {
	log := &model.OperationLog{
		UserID:       userID,
		Username:     username,
		Operation:    operation,
		Resource:     resource,
		ResourceID:   resourceID,
		Method:       method,
		Path:         path,
		RequestBody:  s.truncateString(requestBody, 5000),  // 限制请求体长度
		ResponseBody: s.truncateString(responseBody, 5000), // 限制响应体长度
		IP:           ip,
		UserAgent:    s.truncateString(userAgent, 500), // 限制User-Agent长度
		ResponseCode: responseCode,
		ExecuteTime:  executeTime,
		Status:       status,
		ErrorMessage: s.truncateString(errorMessage, 1000), // 限制错误信息长度
	}

	return s.logDAO.Create(log)
}

// GetLogList 获取操作日志列表
func (s *OperationLogService) GetLogList(req *dto.OperationLogListRequest) (*dto.OperationLogListResponse, error) {
	offset := (req.Page - 1) * req.PageSize

	// 解析时间参数
	var startTime, endTime *time.Time
	if req.StartTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", req.StartTime); err == nil {
			startTime = &t
		}
	}
	if req.EndTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", req.EndTime); err == nil {
			endTime = &t
		}
	}

	logs, total, err := s.logDAO.GetList(offset, req.PageSize, req.UserID, req.Username, req.Operation, req.Resource, req.Method, req.Status, startTime, endTime)
	if err != nil {
		return nil, err
	}

	var logInfos []dto.OperationLogInfo
	for _, log := range logs {
		logInfos = append(logInfos, s.convertToLogInfo(&log))
	}

	return &dto.OperationLogListResponse{
		List:       logInfos,
		Pagination: dto.NewPagination(total, req.Page, req.PageSize),
	}, nil
}

// CleanOldLogs 清理过期日志
func (s *OperationLogService) CleanOldLogs(days int) error {
	return s.logDAO.DeleteOldLogs(days)
}

// convertToLogInfo 转换为LogInfo
func (s *OperationLogService) convertToLogInfo(log *model.OperationLog) dto.OperationLogInfo {
	return dto.OperationLogInfo{
		ID:           log.ID,
		UserID:       log.UserID,
		Username:     log.Username,
		Operation:    log.Operation,
		Resource:     log.Resource,
		ResourceID:   log.ResourceID,
		Method:       log.Method,
		Path:         log.Path,
		RequestBody:  log.RequestBody,
		ResponseCode: log.ResponseCode,
		ResponseBody: log.ResponseBody,
		IP:           log.IP,
		UserAgent:    log.UserAgent,
		ExecuteTime:  log.ExecuteTime,
		Status:       log.Status,
		ErrorMessage: log.ErrorMessage,
		CreatedAt:    log.CreatedAt,
	}
}

// truncateString 截断字符串
func (s *OperationLogService) truncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen] + "..."
}

// LogRequest 记录请求日志的辅助方法
func (s *OperationLogService) LogRequest(userID *uint64, username, operation, resource, resourceID, method, path string, requestBody interface{}, responseCode int, responseBody interface{}, ip, userAgent string, executeTime int64, err error) {
	// 序列化请求体
	var reqBodyStr string
	if requestBody != nil {
		if reqBytes, e := json.Marshal(requestBody); e == nil {
			reqBodyStr = string(reqBytes)
		}
	}

	// 序列化响应体
	var respBodyStr string
	if responseBody != nil {
		if respBytes, e := json.Marshal(responseBody); e == nil {
			respBodyStr = string(respBytes)
		}
	}

	// 确定状态和错误信息
	status := int8(1) // 成功
	errorMessage := ""
	if err != nil {
		status = 0 // 失败
		errorMessage = err.Error()
	}

	// 创建日志
	s.CreateLog(userID, username, operation, resource, resourceID, method, path, reqBodyStr, respBodyStr, ip, userAgent, responseCode, executeTime, status, errorMessage)
}
