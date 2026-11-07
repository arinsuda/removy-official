import axios from "axios";
import { useAuthStore } from "@/stores/auth";
import { resume as resumeSocket } from "@/composables/useSocket";
import { REFRESH_POST_TIMEOUT_MS, isSessionRejected, runExclusiveRefresh } from "@/api/refresh";

const getBaseURL = () => {
  const envUrl = import.meta.env.VITE_API_BASE_URL;
  if (!envUrl) return "/api";
  const cleanUrl = envUrl.replace(/\/+$/, "");
  return cleanUrl.endsWith("/api") ? cleanUrl : `${cleanUrl}/api`;
};

const api = axios.create({
  baseURL: getBaseURL(),
  withCredentials: true,
  headers: {
    "Content-Type": "application/json",
  },
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("access_token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

let isRefreshing = false;

let failedQueue: {
  resolve: (value?: unknown) => void;
  reject: (error?: any) => void;
}[] = [];

function processQueue(error: any = null) {
  failedQueue.forEach((promise) => {
    if (error) {
      promise.reject(error);
    } else {
      promise.resolve();
    }
  });

  failedQueue = [];
}

api.interceptors.response.use(
  (response) => response,

  async (error) => {
    const originalRequest = error.config;
    const status = error.response?.status;

    // ป้องกัน refresh loop — ไม่ retry ถ้าเป็น request จาก auth routes เหล่านี้
    const isAuthRoute = ["/auth/refresh", "/auth/login", "/auth/logout", "/auth/handoff"].some(
      (path) => originalRequest?.url?.includes(path),
    );
    if (isAuthRoute) {
      return Promise.reject(error);
    }

    // retry เฉพาะ 401 เท่านั้น
    if (status !== 401 || originalRequest._retry) {
      return Promise.reject(error);
    }

    originalRequest._retry = true;

    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        failedQueue.push({ resolve, reject });
      }).then(() => api(originalRequest));
    }

    isRefreshing = true;

    try {
      const staleAccess =
        String(originalRequest.headers?.Authorization ?? "").replace(/^Bearer /, "") || null;
      await runExclusiveRefresh(staleAccess, async () => {
        const storedRefresh = localStorage.getItem("refresh_token");
        const refreshUrl = `${getBaseURL()}/auth/refresh`;
        const refreshRes = await axios.post(
          refreshUrl,
          { refresh_token: storedRefresh },
          { withCredentials: true, timeout: REFRESH_POST_TIMEOUT_MS },
        );

        if (refreshRes.data?.access_token) {
          localStorage.setItem("access_token", refreshRes.data.access_token);
        }
        if (refreshRes.data?.refresh_token) {
          localStorage.setItem("refresh_token", refreshRes.data.refresh_token);
        }
        resumeSocket();
      });

      processQueue();

      if (originalRequest.headers) {
        const newAccessToken = localStorage.getItem("access_token");
        if (newAccessToken) {
          originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
        }
      }

      return api(originalRequest);
    } catch (refreshError) {
      processQueue(refreshError);
      // 503 / network error: the server could not check the session. Keep the tokens and
      // the user signed in; the request just fails and can be retried.
      if (!isSessionRejected(refreshError)) {
        return Promise.reject(refreshError);
      }
      localStorage.removeItem("access_token");
      localStorage.removeItem("refresh_token");

      const authStore = useAuthStore();
      authStore.clearUser();
      window.location.href = "/login";

      return Promise.reject(refreshError);
    } finally {
      isRefreshing = false;
    }
  },
);

export default api;
