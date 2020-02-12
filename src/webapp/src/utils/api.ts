/*
 * @Author: Jmeow
 * @Date: 2020-01-28 15:30:50
 * @Description: common API
 */

import Utils from './common';

const utils = new Utils();

const BASE_URL = "api";

class API {
    /**
     * 获取录播机状态
     */
    getLiveInfo() {
        return utils.requestGet(`${BASE_URL}/info`);
    }

    /**
     * 获取直播间列表
     */
    getRoomList() {
        return utils.requestGet(`${BASE_URL}/lives`);
    }

    /**
     * 添加新的直播间
     * @param url URL
     */
    addNewRoom(url: string) {
        const reqBody = [
            {
                "url": url,
                "listen": true
            }
        ];
        return utils.requestPost(`${BASE_URL}/lives`, reqBody);
    }

    /**
     * 删除直播间
     * @param id 直播间id
     */
    deleteRoom(id: string) {
        return utils.requestDelete(`${BASE_URL}/lives/${id}`);
    }

    /**
     * 开始监听直播间
     * @param id 直播间id
     */
    startRecord(id: string) {
        return utils.requestGet(`${BASE_URL}/lives/${id}/start`);
    }

    /**
     * 停止监听直播间
     * @param id 直播间id
     */
    stopRecord(id: string) {
        return utils.requestGet(`${BASE_URL}/lives/${id}/stop`);
    }

    /**
     * 保存设置至config文件
     */
    saveSettings() {
        return utils.requestPut(`${BASE_URL}/config`);
    }

}

export default API;
