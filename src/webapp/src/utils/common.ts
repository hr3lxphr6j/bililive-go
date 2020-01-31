/*
 * @Author: Jmeow
 * @Date: 2020-01-28 11:25:44
 * @Description: common utils
 */

const BASE_URL = "";

class Utils {
    /**
     * Get request
     * @param url URL
     */
    requestGet(url: string) {
        return new Promise((resolve, reject) => {
            fetch(BASE_URL + url)
                .then(rsp => {
                    if (rsp.ok) {
                        return rsp.json();
                    } else {
                        return Promise.reject();
                    }
                }).then(data => {
                    resolve(data);
                }).catch(err => {
                    Utils.alertError();
                    reject(err);
                });
        });
    }

    /**
     * Post request
     * @param url URL
     * @param body Request body
     */
    requestPost(url: string, body?: object) {
        return new Promise((resolve, reject) => {
            fetch(BASE_URL + url, {
                method: 'POST',
                body: JSON.stringify(body),
                headers: new Headers({
                    'Content-Type': 'application/json'
                })
            }).then(rsp => {
                if (rsp.ok) {
                    return rsp.json();
                } else {
                    return Promise.reject();
                }
            }).then(data => {
                resolve(data);
            }).catch(err => {
                Utils.alertError();
                reject(err);
            });
        });
    }

    /**
     * Post request
     * @param url URL
     * @param body Request body
     */
    requestPut(url: string, body?: object) {
        return new Promise((resolve, reject) => {
            fetch(BASE_URL + url, {
                method: 'PUT',
                body: JSON.stringify(body),
                headers: new Headers({
                    'Content-Type': 'application/json'
                })
            }).then(rsp => {
                if (rsp.ok) {
                    return rsp.json();
                } else {
                    return Promise.reject();
                }
            }).then(data => {
                resolve(data);
            }).catch(err => {
                Utils.alertError();
                reject(err);
            });
        });
    }

    /**
     * Delete request
     * @param url URL
     */
    requestDelete(url: string) {
        return new Promise((resolve, reject) => {
            fetch(BASE_URL + url, {
                method: 'DELETE'
            }).then(rsp => {
                if (rsp.ok) {
                    return rsp.json();
                } else {
                    return Promise.reject();
                }
            }).then(data => {
                resolve(data);
            }).catch(err => {
                Utils.alertError();
                reject(err);
            });
        });
    }

    /**
     * Show Error 
     * @param err error Object
     */
    static alertError(err?: any) {
        console.error(err ? err : "Server Error!");
        //alert(err ? err : "Server Error!");
    }
}

export default Utils;