/**
 * 天翼云盘链接解析器
 * 用于解析天翼云盘分享链接和21cn订阅链接，提取分享ID、访问码和uuid
 */

// 解析结果接口
export interface CloudPan189ParseResult {
  shareId: string
  accessCode?: string
  originalUrl: string
  type: 'share' | 'subscribe' // 链接类型
}

// 解析错误类
export class CloudPan189ParseError extends Error {
  constructor(message: string, public originalUrl?: string) {
    super(message)
    this.name = 'CloudPan189ParseError'
  }
}

/**
 * 解析天翼云盘分享链接或21cn订阅链接
 * @param input 输入的链接或包含链接的文本
 * @returns 解析结果
 */
export function parseCloudPan189Link(input: string): CloudPan189ParseResult {
  if (!input || typeof input !== 'string') {
    throw new CloudPan189ParseError('输入不能为空')
  }

  // 清理输入文本
  const cleanInput = input.trim()
  
  // 检查是否为21cn订阅链接
  if (cleanInput.includes('content.21cn.com')) {
    return parse21cnSubscribeLink(cleanInput)
  }
  
  // 解析天翼云盘分享链接
  return parseCloudPanShareLink(cleanInput)
}

/**
 * 解析天翼云盘分享链接
 * @param input 输入的链接或包含链接的文本
 * @returns 解析结果
 */
function parseCloudPanShareLink(input: string): CloudPan189ParseResult {
  // 匹配天翼云盘链接的正则表达式
  // 支持 http 和 https，支持 cloud.189.cn 和 189.cn
  const urlRegex = /https?:\/\/(?:cloud\.)?189\.cn\/t\/([a-zA-Z0-9]+)/g
  
  // 匹配访问码的正则表达式
  // 支持多种格式：访问码：xxxx、密码：xxxx、提取码：xxxx等
  const accessCodeRegex = /(?:访问码|密码|提取码|验证码)[:：]\s*([a-zA-Z0-9]+)/i
  
  const matches = Array.from(input.matchAll(urlRegex))
  
  if (matches.length === 0) {
    throw new CloudPan189ParseError('未找到有效的天翼云盘链接', input)
  }
  
  // 取第一个匹配的链接
  const match = matches[0]
  const shareId = match[1]
  const originalUrl = match[0]
  
  // 查找访问码
  let accessCode: string | undefined
  const accessCodeMatch = input.match(accessCodeRegex)
  if (accessCodeMatch) {
    accessCode = accessCodeMatch[1]
  }
  
  return {
    shareId,
    accessCode,
    originalUrl,
    type: 'share'
  }
}

/**
 * 解析21cn订阅链接
 * @param input 输入的链接
 * @returns 解析结果
 */
function parse21cnSubscribeLink(input: string): CloudPan189ParseResult {
  // 匹配21cn订阅链接中的uuid
  const uuidRegex = /https?:\/\/content\.21cn\.com\/h5\/subscrip\/index\.html#\/pages\/own-home\/index\?uuid=([a-f0-9]{32})/i
  
  const match = input.match(uuidRegex)
  
  if (!match) {
    throw new CloudPan189ParseError('未找到有效的21cn订阅链接或uuid格式错误', input)
  }
  
  const uuid = match[1]
  const originalUrl = match[0]
  
  return {
    shareId: uuid,
    originalUrl,
    type: 'subscribe'
  }
}

/**
 * 批量解析天翼云盘链接和21cn订阅链接
 * @param input 输入的文本
 * @returns 解析结果数组
 */
export function parseMultipleCloudPan189Links(input: string): CloudPan189ParseResult[] {
  if (!input || typeof input !== 'string') {
    return []
  }

  const cleanInput = input.trim()
  const results: CloudPan189ParseResult[] = []
  
  // 匹配天翼云盘链接
  const cloudUrlRegex = /https?:\/\/(?:cloud\.)?189\.cn\/t\/([a-zA-Z0-9]+)/g
  const cloudMatches = Array.from(cleanInput.matchAll(cloudUrlRegex))
  
  for (const match of cloudMatches) {
    const shareId = match[1]
    const originalUrl = match[0]
    
    // 尝试在链接后面查找访问码
    const linkIndex = match.index || 0
    const linkEndIndex = linkIndex + originalUrl.length
    const contextEnd = Math.min(cleanInput.length, linkEndIndex + 100)
    const context = cleanInput.slice(linkEndIndex, contextEnd)
    
    const accessCodeRegex = /(?:访问码|密码|提取码|验证码)[:：]\s*([a-zA-Z0-9]+)/i
    const accessCodeMatch = context.match(accessCodeRegex)
    
    results.push({
      shareId,
      accessCode: accessCodeMatch ? accessCodeMatch[1] : undefined,
      originalUrl,
      type: 'share'
    })
  }
  
  // 匹配21cn订阅链接
  const cnUrlRegex = /https?:\/\/content\.21cn\.com\/h5\/subscrip\/index\.html#\/pages\/own-home\/index\?uuid=([a-f0-9]{32})/gi
  const cnMatches = Array.from(cleanInput.matchAll(cnUrlRegex))
  
  for (const match of cnMatches) {
    const uuid = match[1]
    const originalUrl = match[0]
    
    results.push({
      shareId: uuid,
      originalUrl,
      type: 'subscribe'
    })
  }
  
  return results
}

/**
 * 验证分享ID格式
 * @param shareId 分享ID
 * @param type 链接类型
 * @returns 是否有效
 */
export function isValidShareId(shareId: string, type: 'share' | 'subscribe' = 'share'): boolean {
  if (!shareId || typeof shareId !== 'string') {
    return false
  }
  
  if (type === 'subscribe') {
    // uuid格式：32位十六进制字符
    return /^[a-f0-9]{32}$/i.test(shareId)
  }
  
  // 天翼云盘分享ID通常是字母数字组合，长度在8-20之间
  return /^[a-zA-Z0-9]{8,20}$/.test(shareId)
}

/**
 * 验证访问码格式
 * @param accessCode 访问码
 * @returns 是否有效
 */
export function isValidAccessCode(accessCode: string): boolean {
  if (!accessCode || typeof accessCode !== 'string') {
    return false
  }
  
  // 访问码通常是4位字母数字组合
  return /^[a-zA-Z0-9]{4}$/.test(accessCode)
}

/**
 * 构建天翼云盘分享链接或21cn订阅链接
 * @param shareId 分享ID或uuid
 * @param accessCode 访问码（可选）
 * @param type 链接类型
 * @returns 完整的分享链接
 */
export function buildCloudPan189Link(shareId: string, accessCode?: string, type: 'share' | 'subscribe' = 'share'): string {
  if (!isValidShareId(shareId, type)) {
    throw new CloudPan189ParseError(`无效的${type === 'subscribe' ? 'uuid' : '分享ID'}格式`)
  }
  
  if (type === 'subscribe') {
    return `https://content.21cn.com/h5/subscrip/index.html#/pages/own-home/index?uuid=${shareId}`
  }
  
  const baseUrl = `https://cloud.189.cn/t/${shareId}`
  
  if (accessCode) {
    if (!isValidAccessCode(accessCode)) {
      throw new CloudPan189ParseError('无效的访问码格式')
    }
    return `${baseUrl}（访问码：${accessCode}）`
  }
  
  return baseUrl
}

// 导出默认对象
export default {
  parseCloudPan189Link,
  parseMultipleCloudPan189Links,
  isValidShareId,
  isValidAccessCode,
  buildCloudPan189Link,
  CloudPan189ParseError
}