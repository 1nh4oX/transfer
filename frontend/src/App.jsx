import { useEffect, useRef, useState } from 'react'
import { AlertCircle, CheckCircle2, Clock, FileBox, Folder, LogOut, Send } from 'lucide-react'
import FileArea from './components/file-area'
import { flattenTree, formatTime } from './lib/helpers'

const GLASS_PANEL = 'bg-white/60 backdrop-blur-xl border border-white/60 shadow-[0_8px_30px_rgb(0,0,0,0.04)]'
const SPRING_TRANSITION = 'transition-all duration-300 ease-out'
const TOKEN_KEY = 'cloud_token'
const AUTH_EXPIRED_EVENT = 'auth-expired'
const API_BASE = import.meta.env.VITE_API_BASE || ''
const USER_QUOTA_BYTES = 2 * 1024 * 1024 * 1024

async function apiFetch(path, options = {}, token) {
  const headers = new Headers(options.headers || {})
  if (!headers.has('Content-Type') && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }
  if (token) headers.set('Authorization', `Bearer ${token}`)

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers })
  const contentType = res.headers.get('content-type') || ''
  const body = contentType.includes('application/json') ? await res.json().catch(() => ({})) : null

  if (!res.ok) {
    if (res.status === 401) {
      localStorage.removeItem(TOKEN_KEY)
      window.dispatchEvent(new CustomEvent(AUTH_EXPIRED_EVENT))
    }
    const msg = body?.message || `Request failed: ${res.status}`
    throw new Error(msg)
  }
  return body
}

function Toast({ message, type }) {
  if (!message) return null
  return (
    <div
      className={`fixed top-6 left-1/2 -translate-x-1/2 z-50 px-6 py-3 rounded-full flex items-center gap-2 shadow-lg border ${
        type === 'error' ? 'bg-red-50 border-red-200 text-red-600' : 'bg-slate-800 text-white border-slate-700'
      }`}
    >
      {type === 'error' ? <AlertCircle size={16} /> : <CheckCircle2 size={16} />}
      <span className="text-sm font-medium">{message}</span>
    </div>
  )
}

function LoginPage({ onLogin, showToast }) {
  const [mode, setMode] = useState('login')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [errorMessage, setErrorMessage] = useState('')

  const submit = async (e) => {
    e.preventDefault()
    setErrorMessage('')
    const cleanUsername = username.trim()
    if (!cleanUsername || !password) {
      const msg = '请输入用户名和密码'
      setErrorMessage(msg)
      showToast(msg, 'error')
      return
    }
    if (mode === 'register' && password !== confirmPassword) {
      setErrorMessage('两次密码不一致')
      showToast('两次密码不一致', 'error')
      return
    }
    setLoading(true)
    try {
      const path = mode === 'register' ? '/api/auth/register' : '/api/auth/login'
      const data = await apiFetch(path, {
        method: 'POST',
        body: JSON.stringify({ username: cleanUsername, password }),
      })
      if (!data?.token || typeof data.token !== 'string') {
        throw new Error('登录接口返回异常：缺少 token')
      }
      onLogin(data.token)
      showToast(mode === 'register' ? '注册并登录成功' : '登录成功')
    } catch (err) {
      const msg = err instanceof Error ? err.message : '登录/注册失败'
      setErrorMessage(msg)
      showToast(msg, 'error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-[#f4f5f7] flex items-center justify-center p-4">
      <div className={`relative w-full max-w-md ${GLASS_PANEL} p-10 rounded-[2rem]`}>
        <div className="mb-8 text-center">
          <div className="w-12 h-12 bg-slate-800 rounded-2xl flex items-center justify-center text-white mx-auto mb-4">
            <FileBox size={24} />
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-slate-800">登录 CloudSpace</h1>
          <p className="text-sm text-slate-500 mt-2">输入账号密码即可登录或注册</p>
        </div>

        <div className="mb-4 flex bg-slate-100 rounded-xl p-1">
          <button
            type="button"
            onClick={() => setMode('login')}
            className={`flex-1 py-2 text-sm rounded-lg ${mode === 'login' ? 'bg-white shadow-sm' : 'text-slate-500'}`}
          >
            登录
          </button>
          <button
            type="button"
            onClick={() => setMode('register')}
            className={`flex-1 py-2 text-sm rounded-lg ${mode === 'register' ? 'bg-white shadow-sm' : 'text-slate-500'}`}
          >
            注册
          </button>
        </div>

        <form onSubmit={submit} className="space-y-5">
          {errorMessage && (
            <div className="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600">{errorMessage}</div>
          )}
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            className="w-full bg-white/70 border border-white/80 px-4 py-3 rounded-xl text-sm focus:outline-none"
            placeholder="用户名"
          />
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full bg-white/70 border border-white/80 px-4 py-3 rounded-xl text-sm focus:outline-none"
            placeholder="密码"
          />
          {mode === 'register' && (
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="w-full bg-white/70 border border-white/80 px-4 py-3 rounded-xl text-sm focus:outline-none"
              placeholder="确认密码"
            />
          )}
          <button
            type="submit"
            disabled={loading}
            className="w-full bg-slate-800 text-white font-medium py-3 rounded-xl disabled:opacity-50"
          >
            {loading ? (mode === 'register' ? '注册中...' : '登录中...') : mode === 'register' ? '注册并进入' : '进入空间'}
          </button>
        </form>
      </div>
    </div>
  )
}

function Sidebar({ tree, currentFolderID, onSelectFolder, mobile = false }) {
  const renderNodes = (nodes, depth = 0) => {
    return (nodes || []).map((node) => {
      const active = currentFolderID === node.folder.id
      return (
        <div key={node.folder.id}>
          <button
            onClick={() => onSelectFolder(node.folder.id)}
            className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm text-left ${
              active ? 'bg-white shadow-sm text-slate-800 font-medium' : 'text-slate-600 hover:bg-white/70'
            }`}
            style={{ paddingLeft: `${12 + depth * 16}px` }}
          >
            <Folder size={16} className="text-slate-600 shrink-0" />
            <span className="truncate">{node.folder.name}</span>
          </button>
          {node.children?.length > 0 && renderNodes(node.children, depth + 1)}
        </div>
      )
    })
  }

  return (
    <aside className={`${mobile ? 'w-full h-full' : 'hidden lg:flex w-64 shrink-0'} flex flex-col ${GLASS_PANEL} rounded-3xl p-5`}>
      <h3 className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-4 px-2">目录</h3>
      <button
        onClick={() => onSelectFolder(null)}
        className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm mb-2 ${
          currentFolderID == null ? 'font-medium bg-white shadow-sm text-slate-800' : 'text-slate-600 hover:bg-white/70'
        }`}
      >
        <Folder size={18} className="text-slate-700" />
        根目录
      </button>
      <div className="flex-1 overflow-auto space-y-1 pr-1">{renderNodes(tree)}</div>
    </aside>
  )
}

function NotesArea({ token, showToast, mobile = false }) {
  const [notes, setNotes] = useState([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const scrollRef = useRef(null)

  const loadNotes = async () => {
    try {
      const data = await apiFetch('/api/notes?page=1&pageSize=100', {}, token)
      const items = data.items || []
      setNotes(items.slice().reverse())
    } catch (err) {
      showToast(err.message, 'error')
    }
  }

  useEffect(() => {
    loadNotes()
  }, [])

  useEffect(() => {
    if (scrollRef.current) scrollRef.current.scrollTop = scrollRef.current.scrollHeight
  }, [notes])

  const send = async () => {
    const content = input.trim()
    if (!content) return
    setLoading(true)
    setInput('')
    try {
      const item = await apiFetch('/api/notes', { method: 'POST', body: JSON.stringify({ content }) }, token)
      setNotes((prev) => [...prev, item])
    } catch (err) {
      showToast(err.message, 'error')
      setInput(content)
    } finally {
      setLoading(false)
    }
  }

  return (
    <aside className={`${mobile ? 'w-full h-full' : 'w-[360px] h-full shrink-0'} min-h-0 flex flex-col ${GLASS_PANEL} rounded-3xl overflow-hidden`}>
      <div className="p-4 border-b border-white/50 bg-white/30 flex items-center justify-between">
        <div>
          <h3 className="font-semibold text-sm text-slate-800 tracking-wide">速记对话框</h3>
          <p className="text-[10px] text-slate-500">仅自己可见</p>
        </div>
      </div>

      <div className="flex-1 min-h-0 overflow-y-auto overscroll-contain p-4 space-y-5 flex flex-col bg-slate-50/30" ref={scrollRef}>
        {notes.length === 0 ? (
          <div className="flex-1 flex flex-col items-center justify-center text-slate-400 mt-10 opacity-60">
            <Clock size={32} className="mb-2 text-slate-300" />
            <span className="text-xs">在这里随手记录想法或待办...</span>
          </div>
        ) : (
          notes.map((note) => (
            <div key={note.id} className="flex flex-col items-end">
              <div className="bg-slate-800 text-white p-3.5 rounded-2xl rounded-br-sm shadow-md max-w-[85%]">
                <p className="text-sm leading-relaxed whitespace-pre-wrap break-words">{note.content}</p>
              </div>
              <span className="text-[10px] text-slate-400 mt-1.5 mr-1 font-medium">{formatTime(note.createdAt)}</span>
            </div>
          ))
        )}
        {loading && <div className="text-xs text-slate-400 text-right">发送中...</div>}
      </div>

      <div className="p-4 bg-white/40 border-t border-white/50">
        <div className="relative flex items-end bg-white/80 border border-slate-200/60 rounded-[20px] shadow-sm">
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                send()
              }
            }}
            placeholder="随时记录..."
            className="w-full bg-transparent rounded-[20px] pl-4 pr-12 py-3 text-sm focus:outline-none resize-none h-[44px]"
            rows="1"
          />
          <button
            onClick={send}
            disabled={!input.trim()}
            className={`absolute right-2 bottom-2 p-2 rounded-xl flex items-center justify-center ${
              input.trim() ? 'bg-slate-800 text-white' : 'bg-slate-100 text-slate-400 cursor-not-allowed'
            }`}
          >
            <Send size={16} />
          </button>
        </div>
      </div>
    </aside>
  )
}

export default function App() {
  const [token, setToken] = useState(localStorage.getItem(TOKEN_KEY) || '')
  const [toast, setToast] = useState({ message: '', type: '' })
  const [tree, setTree] = useState([])
  const [currentFolderID, setCurrentFolderID] = useState(null)
  const [totalUsedBytes, setTotalUsedBytes] = useState(0)
  const [mobileView, setMobileView] = useState('files')

  const showToast = (message, type = 'success') => {
    setToast({ message, type })
    window.clearTimeout(window.__toastTimer)
    window.__toastTimer = window.setTimeout(() => setToast({ message: '', type: '' }), 3000)
  }

  const logout = () => {
    localStorage.removeItem(TOKEN_KEY)
    setToken('')
  }

  const setLoggedIn = (nextToken) => {
    localStorage.setItem(TOKEN_KEY, nextToken)
    setToken(nextToken)
    setCurrentFolderID(null)
    setMobileView('files')
  }

  const loadTree = async () => {
    if (!token) return
    try {
      const data = await apiFetch('/api/tree', {}, token)
      setTree(data.rootFolders || [])
    } catch (err) {
      showToast(err.message, 'error')
    }
  }

  useEffect(() => {
    loadTree()
  }, [token])

  useEffect(() => {
    const onAuthExpired = () => {
      if (token) {
        setToken('')
        showToast('登录已失效，请重新登录', 'error')
      }
    }
    window.addEventListener(AUTH_EXPIRED_EVENT, onAuthExpired)
    return () => window.removeEventListener(AUTH_EXPIRED_EVENT, onAuthExpired)
  }, [token])

  const loadTotalUsage = async () => {
    if (!token) return
    try {
      const folderIDs = [null, ...flattenTree(tree).map((f) => f.id)]
      let sum = 0
      for (const fid of folderIDs) {
        let page = 1
        while (true) {
          const query = fid
            ? `/api/files?page=${page}&pageSize=100&folderId=${encodeURIComponent(fid)}`
            : `/api/files?page=${page}&pageSize=100`
          const data = await apiFetch(query, {}, token)
          const items = data.items || []
          sum += items.reduce((acc, item) => acc + (item.size || 0), 0)
          const total = data.total || 0
          if (page * 100 >= total || items.length === 0) break
          page += 1
        }
      }
      setTotalUsedBytes(sum)
    } catch (err) {
      showToast(err.message, 'error')
    }
  }

  useEffect(() => {
    loadTotalUsage()
  }, [token, tree])

  if (!token) {
    return (
      <div className="min-h-dvh bg-[#f4f5f7]">
        <Toast message={toast.message} type={toast.type} />
        <LoginPage onLogin={setLoggedIn} showToast={showToast} />
      </div>
    )
  }

  return (
    <div className="h-dvh min-h-dvh bg-[#f4f5f7] text-slate-800 font-sans overflow-hidden flex flex-col">
      <Toast message={toast.message} type={toast.type} />

      <header className={`h-16 ${GLASS_PANEL} rounded-b-2xl mx-3 sm:mx-4 mt-2 flex items-center justify-between px-4 sm:px-6 z-10 shrink-0`}>
        <div className="flex items-center gap-3 min-w-0">
          <div className="w-8 h-8 rounded-lg bg-slate-800 flex items-center justify-center text-white shadow-sm shrink-0">
            <FileBox size={18} />
          </div>
          <span className="font-semibold text-base sm:text-lg tracking-tight truncate">Mono&apos;s CloudSpace</span>
        </div>
        <button onClick={logout} className={`p-2 rounded-full hover:bg-slate-200/50 text-slate-500 hover:text-slate-800 ${SPRING_TRANSITION}`}>
          <LogOut size={18} />
        </button>
      </header>

      <main className="hidden lg:flex flex-1 min-h-0 overflow-hidden p-4 gap-4 max-w-[1600px] mx-auto w-full">
        <Sidebar tree={tree} currentFolderID={currentFolderID} onSelectFolder={setCurrentFolderID} />
        <FileArea
          token={token}
          showToast={showToast}
          apiFetch={apiFetch}
          tree={tree}
          currentFolderID={currentFolderID}
          onSelectFolder={setCurrentFolderID}
          refreshTree={loadTree}
          totalUsedBytes={totalUsedBytes}
        />
        <NotesArea token={token} showToast={showToast} />
      </main>

      <main className="lg:hidden flex-1 min-h-0 overflow-hidden p-3 pb-[env(safe-area-inset-bottom)] flex flex-col gap-3">
        <div className={`${GLASS_PANEL} rounded-2xl p-1 grid grid-cols-3 gap-1 shrink-0`}>
          <button
            onClick={() => setMobileView('files')}
            className={`py-2 rounded-xl text-sm font-medium ${mobileView === 'files' ? 'bg-white shadow-sm text-slate-800' : 'text-slate-500'}`}
          >
            Files
          </button>
          <button
            onClick={() => setMobileView('notes')}
            className={`py-2 rounded-xl text-sm font-medium ${mobileView === 'notes' ? 'bg-white shadow-sm text-slate-800' : 'text-slate-500'}`}
          >
            Notes
          </button>
          <button
            onClick={() => setMobileView('folders')}
            className={`py-2 rounded-xl text-sm font-medium ${mobileView === 'folders' ? 'bg-white shadow-sm text-slate-800' : 'text-slate-500'}`}
          >
            Folders
          </button>
        </div>

        <div className="flex-1 min-h-0 overflow-hidden">
          {mobileView === 'files' && (
            <FileArea
              mobile
              token={token}
              showToast={showToast}
              apiFetch={apiFetch}
              tree={tree}
              currentFolderID={currentFolderID}
              onSelectFolder={setCurrentFolderID}
              refreshTree={loadTree}
              totalUsedBytes={totalUsedBytes}
            />
          )}
          {mobileView === 'notes' && <NotesArea mobile token={token} showToast={showToast} />}
          {mobileView === 'folders' && (
            <Sidebar mobile tree={tree} currentFolderID={currentFolderID} onSelectFolder={setCurrentFolderID} />
          )}
        </div>
      </main>
    </div>
  )
}
