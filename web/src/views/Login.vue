<template>
  <div class="login-container">
    <div class="login-background">
      <div class="bg-circle circle-1" />
      <div class="bg-circle circle-2" />
      <div class="bg-circle circle-3" />
    </div>
    <div class="login-box">
      <div class="login-header">
        <div class="logo-wrapper">
          <div class="logo-icon">
            <svg viewBox="0 0 1147 1024" height="60">
              <path d="M577.980145 584.790361L955.058892 766.026024v38.640578l-377.078747 209.981687-377.06641-218.95094v-34.149783z" fill="#1890FF" opacity=".2" p-id="7886" />
              <path
                d="M579.238554 978.15441L959.969157 766.13706 579.324916 531.863133l-2.553832-0.024675-380.69359 225.23065 380.718265 221.060627 2.430458 0.024675zM950.222651 765.902651L578.066506 973.157783 205.836337 757.019759 578.004819 536.847422 950.222651 765.902651z"
                fill="#FFFFFF"
                opacity=".2"
                p-id="7887"
              />
              <path d="M577.77041 302.857253l502.771662 243.243181v43.846939L577.758072 871.732434 75.023422 577.90612v-37.826313z" fill="#1890FF" opacity=".4" p-id="7888" />
              <path
                d="M579.028819 830.007518L1085.44 546.186795l-506.324819-313.615422-2.553832-0.024674L70.199518 534.046843l506.374169 295.936 2.455132 0.024675z m496.701687-284.030458L577.844434 824.998554 79.909012 533.997494l497.886072-296.454169 497.935422 308.433735z"
                fill="#FFFFFF"
                opacity=".4"
                p-id="7889"
              />
              <path d="M571.515373 89.988627l565.605784 270.558072V412.06747L571.515373 725.497831 5.921928 398.656771v-44.809253z" fill="#1890FF" opacity=".6" p-id="7890" />
              <path
                d="M572.761446 676.123759L1142.068434 360.645398 572.847807 12.028916l-2.541494-0.024675L1.061012 347.148337l569.269976 328.963085 2.430458 0.012337z m559.52347-315.700434L571.589398 671.127133 10.856867 347.098988 571.540048 17.025542l560.73253 343.410121z"
                fill="#FFFFFF"
                opacity=".6"
                p-id="7891"
              />
            </svg>
          </div>
        </div>
        <h1 class="login-title">YSIP 管理平台</h1>
        <p class="login-subtitle">GB28181 SIP 服务器</p>
      </div>

      <div class="login-form-wrapper">
        <a-form :model="formState" layout="vertical" @finish="handleLogin">
          <a-form-item name="username" :rules="[{ required: true, message: '请输入用户名' }]">
            <a-input v-model:value="formState.username" placeholder="请输入用户名" size="large" class="custom-input">
              <template #prefix>
                <UserOutlined />
              </template>
            </a-input>
          </a-form-item>

          <a-form-item name="password" :rules="[{ required: true, message: '请输入密码' }]">
            <a-input-password v-model:value="formState.password" placeholder="请输入密码" size="large" class="custom-input">
              <template #prefix>
                <LockOutlined />
              </template>
            </a-input-password>
          </a-form-item>

          <a-form-item>
            <a-button type="primary" html-type="submit" size="large" block :loading="loading" class="login-button">
              <span v-if="!loading">登 录</span>
              <span v-else>登录中...</span>
            </a-button>
          </a-form-item>
        </a-form>
      </div>

      <div class="login-footer">
        <p class="copyright">© 2026 YoSIP Team. All rights reserved.</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { UserOutlined, LockOutlined } from '@ant-design/icons-vue'
import { useUserStore } from '@/store/user'

const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)
const formState = reactive({
  username: '',
  password: ''
})

const handleLogin = async () => {
  loading.value = true
  try {
    await userStore.login({
      username: formState.username,
      password: formState.password
    })
    message.success('登录成功')
    // 登录成功后跳转到首页
    router.push('/')
  } catch (error) {
    // 错误已在 request 拦截器中处理
  } finally {
    loading.value = false
  }
}
</script>

<style lang="less" scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  position: relative;
  overflow: hidden;

  .login-background {
    position: absolute;
    width: 100%;
    height: 100%;
    overflow: hidden;
    z-index: 0;

    .bg-circle {
      position: absolute;
      border-radius: 50%;
      background: rgba(255, 255, 255, 0.1);
      animation: float 20s infinite ease-in-out;

      &.circle-1 {
        width: 300px;
        height: 300px;
        top: -100px;
        left: -100px;
        animation-delay: 0s;
      }

      &.circle-2 {
        width: 400px;
        height: 400px;
        bottom: -150px;
        right: -150px;
        animation-delay: 5s;
      }

      &.circle-3 {
        width: 200px;
        height: 200px;
        top: 50%;
        right: 10%;
        animation-delay: 10s;
      }
    }

    @keyframes float {
      0%,
      100% {
        transform: translate(0, 0) scale(1);
      }
      33% {
        transform: translate(30px, -30px) scale(1.1);
      }
      66% {
        transform: translate(-20px, 20px) scale(0.9);
      }
    }
  }

  .login-box {
    position: relative;
    z-index: 1;
    width: 420px;
    padding: 48px 40px;
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(10px);
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
    animation: slideUp 0.6s ease-out;

    @keyframes slideUp {
      from {
        opacity: 0;
        transform: translateY(30px);
      }
      to {
        opacity: 1;
        transform: translateY(0);
      }
    }

    .login-header {
      text-align: center;
      margin-bottom: 40px;

      .logo-wrapper {
        display: flex;
        justify-content: center;
        margin-bottom: 20px;

        .logo-icon {
          animation: logoPulse 2s infinite ease-in-out;

          @keyframes logoPulse {
            0%,
            100% {
              transform: scale(1);
            }
            50% {
              transform: scale(1.05);
            }
          }
        }
      }

      .login-title {
        font-size: 28px;
        font-weight: 600;
        color: #1a1a2e;
        margin: 0 0 8px 0;
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
      }

      .login-subtitle {
        font-size: 14px;
        color: #888;
        margin: 0;
        font-weight: 400;
      }
    }

    .login-form-wrapper {
      :deep(.ant-form-item) {
        margin-bottom: 24px;
      }

      .custom-input {
        border-radius: 0px !important;
        :deep(.ant-input) {
          height: 28px;
          font-size: 15px;
          border-radius: 0px;
          transition: all 0.3s ease;
          background: rgba(255, 255, 255, 0.9);

          &:hover {
            border-color: #667eea;
          }

          &:focus {
            border-color: #667eea;
            box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.15);
          }

          &::placeholder {
            color: #bbb;
          }
        }

        :deep(.ant-input-prefix) {
          color: #aaa;
          transition: color 0.3s ease;
          margin-right: 8px;

          svg {
            font-size: 18px;
          }
        }

        &:hover {
          :deep(.ant-input-prefix) {
            color: #667eea;
          }
        }

        :deep(.ant-input:focus) + .ant-input-prefix,
        :deep(.ant-input-affix-wrapper-focused) .ant-input-prefix {
          color: #667eea;
        }
      }

      .login-button {
        height: 48px;
        font-size: 16px;
        font-weight: 500;
        border-radius: 0px;
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        border: none;
        transition: all 0.3s ease;
        margin-top: 12px;

        &:hover {
          transform: translateY(-2px);
          box-shadow: 0 8px 20px rgba(102, 126, 234, 0.4);
        }

        &:active {
          transform: translateY(0);
        }
      }
    }

    .login-footer {
      margin-top: 32px;
      text-align: center;

      .copyright {
        font-size: 12px;
        color: #999;
        margin: 0;
      }
    }
  }
}

// 响应式设计
@media (max-width: 768px) {
  .login-container {
    padding: 20px;

    .login-box {
      width: 100%;
      max-width: 400px;
      padding: 36px 24px;
    }
  }
}
</style>
