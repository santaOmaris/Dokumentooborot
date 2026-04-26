import axios from 'axios'

const api = axios.create({
  baseURL: '/',
  withCredentials: true, // JWT живёт в httpOnly cookie
})

export default api
