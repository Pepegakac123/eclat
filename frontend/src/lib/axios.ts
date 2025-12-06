import { API_BASE_URL } from "@/config/constants";
import axios from "axios";
const apiReq = axios.create({
  baseURL: `${API_BASE_URL}/api`,
  headers: { "Content-Type": "application/json" },
  timeout: 10000,
  paramsSerializer: {
    indexes: null,
  },
});

apiReq.interceptors.response.use(
  (response) => response,
  (error) => {
    const message = error.response?.data?.message || error.message;
    console.error("🔥 API Error:", message);
    // TODO: Tutaj w przyszłości wstawimy: toast.error(message);
    return Promise.reject(error);
  },
);

export default apiReq;
