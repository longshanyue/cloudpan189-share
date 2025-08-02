// 权限位常量定义
export const PERMISSIONS = {
  BASE: 1,      // 基础用户权限
  DAV_READ: 2,  // WebDAV访问权限
  ADMIN: 4      // 管理员权限
} as const

// 权限标签映射
export const PERMISSION_LABELS = {
  [PERMISSIONS.BASE]: '基础权限',
  [PERMISSIONS.DAV_READ]: 'WebDAV访问',
  [PERMISSIONS.ADMIN]: '管理员权限'
} as const

// 权限描述映射
export const PERMISSION_DESCRIPTIONS = {
  [PERMISSIONS.BASE]: '基本的系统访问权限，包括查看个人信息、修改密码等',
  [PERMISSIONS.DAV_READ]: '通过WebDAV协议访问文件系统的权限',
  [PERMISSIONS.ADMIN]: '系统管理权限，包括用户管理、系统设置等'
} as const

/**
 * 获取权限标签列表
 * @param permissions 权限位值
 * @returns 权限标签数组
 */
export function getPermissionLabels(permissions: number): string[] {
  const labels: string[] = []
  
  if (permissions & PERMISSIONS.BASE) {
    labels.push(PERMISSION_LABELS[PERMISSIONS.BASE])
  }
  if (permissions & PERMISSIONS.DAV_READ) {
    labels.push(PERMISSION_LABELS[PERMISSIONS.DAV_READ])
  }
  if (permissions & PERMISSIONS.ADMIN) {
    labels.push(PERMISSION_LABELS[PERMISSIONS.ADMIN])
  }
  
  return labels
}

/**
 * 获取权限详细信息列表
 * @param permissions 权限位值
 * @returns 权限详细信息数组
 */
export function getPermissionDetails(permissions: number): Array<{label: string, description: string, value: number}> {
  const details: Array<{label: string, description: string, value: number}> = []
  
  if (permissions & PERMISSIONS.BASE) {
    details.push({
      label: PERMISSION_LABELS[PERMISSIONS.BASE],
      description: PERMISSION_DESCRIPTIONS[PERMISSIONS.BASE],
      value: PERMISSIONS.BASE
    })
  }
  if (permissions & PERMISSIONS.DAV_READ) {
    details.push({
      label: PERMISSION_LABELS[PERMISSIONS.DAV_READ],
      description: PERMISSION_DESCRIPTIONS[PERMISSIONS.DAV_READ],
      value: PERMISSIONS.DAV_READ
    })
  }
  if (permissions & PERMISSIONS.ADMIN) {
    details.push({
      label: PERMISSION_LABELS[PERMISSIONS.ADMIN],
      description: PERMISSION_DESCRIPTIONS[PERMISSIONS.ADMIN],
      value: PERMISSIONS.ADMIN
    })
  }
  
  return details
}

/**
 * 计算权限位值
 * @param permissionObj 权限对象
 * @returns 权限位值
 */
export function calculatePermissions(permissionObj: {
  base?: boolean
  davRead?: boolean
  admin?: boolean
}): number {
  let permissions = 0
  if (permissionObj.base) permissions |= PERMISSIONS.BASE
  if (permissionObj.davRead) permissions |= PERMISSIONS.DAV_READ
  if (permissionObj.admin) permissions |= PERMISSIONS.ADMIN
  return permissions
}

/**
 * 解析权限位为对象
 * @param permissions 权限位值
 * @returns 权限对象
 */
export function parsePermissions(permissions: number): {
  base: boolean
  davRead: boolean
  admin: boolean
} {
  return {
    base: !!(permissions & PERMISSIONS.BASE),
    davRead: !!(permissions & PERMISSIONS.DAV_READ),
    admin: !!(permissions & PERMISSIONS.ADMIN)
  }
}

/**
 * 检查是否有指定权限
 * @param userPermissions 用户权限位值
 * @param requiredPermission 需要检查的权限
 * @returns 是否有权限
 */
export function hasPermission(userPermissions: number, requiredPermission: number): boolean {
  return (userPermissions & requiredPermission) !== 0
}

/**
 * 检查是否为管理员
 * @param permissions 权限位值
 * @returns 是否为管理员
 */
export function isAdmin(permissions: number): boolean {
  return hasPermission(permissions, PERMISSIONS.ADMIN)
}