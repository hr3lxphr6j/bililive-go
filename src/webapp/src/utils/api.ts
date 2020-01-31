/*
 * @Author: Jmeow
 * @Date: 2020-01-28 15:30:50
 * @Description: common API
 */

import Utils from './common';

const utils = new Utils();

class API {
    /**
     * 获取录播机状态
     */
    getLiveInfo(){
        return utils.requestGet("/info");
    }

    /**
     * 获取直播间列表
     */
    getRoomList() {
        return utils.requestGet("/lives");
    }

    /**
     * 添加新的直播间
     * @param url URL
     */
    addNewRoom(url: string) {
        const reqBody = {
            lives: [
                {
                    "url": url,
                    "listen": true
                }
            ]
        };
        return utils.requestPost("/lives", reqBody);
    }

    /**
     * 删除直播间
     * @param id 直播间id
     */
    deleteRoom(id: string) {
        return utils.requestDelete(`/lives/${id}`);
    }

    /**
     * 开始监听直播间
     * @param id 直播间id
     */
    startRecord(id: string) {
        return utils.requestGet(`/lives/${id}/start`);
    }

    /**
     * 停止监听直播间
     * @param id 直播间id
     */
    stopRecord(id: string) {
        return utils.requestGet(`/lives/${id}/stop`);
    }

    /**
     * 保存设置至config文件
     */
    saveSettings(){
        return utils.requestPut("/config");
    }

}

export default API;