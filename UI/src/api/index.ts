// API 客户端 barrel：所有调用方统一 `import { xxxApi, type YyyItem } from "@/api"`
// 内部按域拆到 ./common ./auth ./user ./menu ./role ./organization ./tenant
//      ./dict ./config ./resource ./frame ./frameCategory ./avatar
//      ./avatarCategory ./space ./asset ./system

export {
  api,
  setAuthTokens,
  clearAuthTokens,
  getToken,
  getRefreshToken,
  ApiError,
  type ApiOptions,
  type ApiResponse,
  type PageResponse,
} from "./common"

export {
  authApi,
  type LoginRequest,
  type LoginResponse,
  type RegisterRequest,
  type RefreshRequest,
  type RefreshResponse,
} from "./auth"

export {
  userApi,
  type UserItem,
} from "./user"

export {
  menuApi,
  type MenuItem,
} from "./menu"

export {
  roleApi,
  type RoleItem,
} from "./role"

export {
  organizationApi,
  type OrganizationItem,
} from "./organization"

export {
  resourceApi,
  type ResourceItem,
} from "./resource"

export {
  tenantApi,
  type TenantItem,
} from "./tenant"

export {
  dictApi,
  type DictItem,
  type DictValueItem,
} from "./dict"

export {
  configApi,
  type ConfigItemType,
  type ConfigOption,
  type ConfigValidation,
  type ConfigGroup,
  type ConfigItem,
  type PublicConfigResponse,
} from "./config"

export {
  frameApi,
  type FrameItem,
} from "./frame"

export {
  frameCategoryApi,
  type FrameCategoryItem,
} from "./frameCategory"

export {
  avatarApi,
  type AvatarItem,
} from "./avatar"

export {
  avatarCategoryApi,
  type AvatarCategoryItem,
} from "./avatarCategory"

export {
  spaceApi,
  type SpaceItem,
  type GenerateAvatarResponse,
} from "./space"

export {
  assetApi,
  type AssetUploadResponse,
} from "./asset"

export {
  systemApi,
  type CacheInfo,
  type CacheKeyItem,
  type CacheValue,
} from "./system"
