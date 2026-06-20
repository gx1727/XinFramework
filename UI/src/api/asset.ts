// 通用资产上传（不走 api()，直接 fetch）

import { API_BASE_URL, getToken, ApiError, type ApiResponse } from "./common"

export interface AssetUploadResponse {
  id: number
  url: string
}

export const assetApi = {
  upload: (file: File) => {
    const formData = new FormData()
    formData.append("file", file)
    const token = getToken()

    const headers: HeadersInit = {}
    if (token) {
      (headers as Record<string, string>)["Authorization"] = `Bearer ${token}`
    }

    return fetch(`${API_BASE_URL}/asset/upload`, {
      method: "POST",
      headers,
      body: formData,
    }).then(async (response) => {
      const data = await response.json()
      if (!response.ok) {
        throw new ApiError(
          response.status,
          data?.code || response.status,
          data?.msg || `Upload failed: ${response.status}`,
          data
        )
      }
      const apiResponse = data as ApiResponse<AssetUploadResponse>
      if (apiResponse.code !== 0) {
        throw new ApiError(200, apiResponse.code, apiResponse.msg, apiResponse.data)
      }
      return apiResponse.data as AssetUploadResponse
    })
  },
}
