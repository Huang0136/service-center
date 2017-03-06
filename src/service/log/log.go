package log

import (
	"beans"
	"constants"
	"log"
	"time"
)

// 保存请求日志
func SaveReqLog(reqLog beans.RequestLog) {
	tx, err := constants.ScDB.Begin()
	if err != nil {
		log.Println("开启事务失败,", err)
	}

	t1 := time.Now()
	strSql := `insert into request_log(ID,REQUEST_IP,OP_CHANNEL,ACCESS_USER,SERVICE_CODE,IN_PARAMS,UC_RECEIVE_TIME,
				UC_RETURN_TIME,CALL_NODE_COUNT,DEAL_RESULT,REQUEST_DEAL_STATUS)
				values(?,?,?,?,?,?,?,?,?,?,?)`
	stat, err := tx.Prepare(strSql)
	if err != nil {
		log.Println("预编译SQL出错,", err)
	}

	rs, err := stat.Exec(reqLog.Id, reqLog.RequestIP, reqLog.OpChannel, reqLog.AccessUser, reqLog.ServiceCode, reqLog.InParams, reqLog.ScReceiveTime, reqLog.ScReturnTime, reqLog.CallNodeCount, reqLog.DealResult, reqLog.RequestDealStatus)
	if err != nil {
		log.Println("执行SQL出错,", err)
	}

	lastId, _ := rs.LastInsertId()
	rowsAffected, _ := rs.RowsAffected()
	tx.Commit()

	t2 := time.Now()
	log.Println("写入日志成功,", lastId, ",", rowsAffected, ",耗时:s", t2.Sub(t1).Seconds())
}
