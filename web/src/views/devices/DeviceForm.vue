<template>
  <a-modal
    v-model:open="visible"
    :title="type === 'add' ? '添加设备' : '设备信息'"
    :footer="null"
    width="500px"
    @cancel="handleCancel"
  >
    <a-form
      :model="form"
      layout="vertical"
      @submit="handleSubmit"
    >
      <a-form-item
        v-if="type === 'edit'"
        label="设备ID"
        name="deviceid"
      >
        <a-input
          v-model:value="form.deviceid"
          disabled
        />
      </a-form-item>
      <a-form-item
        label="设备名称"
        name="name"
        :rules="[{ required: true, message: '请输入设备名称' }]"
      >
        <a-input
          v-model:value="form.name"
          placeholder="请输入设备名称"
        />
      </a-form-item>
      <template v-if="type === 'edit'">
        <a-form-item
          label="收流地址"
          name="host"
        >
          <a-input
            v-model:value="form.host"
            placeholder="留空使用默认地址，如 192.168.1.100"
          />
        </a-form-item>
        <a-form-item
          label="厂商"
          name="manufacturer"
        >
          <a-input
            v-model:value="form.manufacturer"
            placeholder="请输入厂商"
          />
        </a-form-item>
        <a-form-item
          label="型号"
          name="model"
        >
          <a-input
            v-model:value="form.model"
            placeholder="请输入型号"
          />
        </a-form-item>
        <a-form-item
          label="设备密码"
          name="pwd"
        >
          <a-input-password
            v-model:value="form.pwd"
            placeholder="留空则不修改"
          />
        </a-form-item>
      </template>
      <template v-else>
        <a-form-item
          label="设备ID"
          name="deviceId"
        >
          <a-input
            v-model:value="form.deviceId"
            placeholder="不填则自动生成"
          />
        </a-form-item>
        <a-form-item
          label="设备密码"
          name="pwd"
          :rules="[{ required: true, message: '请输入设备密码' }]"
        >
          <a-input-password
            v-model:value="form.pwd"
            placeholder="请输入设备密码"
          />
        </a-form-item>
      </template>
      <a-form-item>
        <a-space>
          <a-button
            type="primary"
            html-type="submit"
            :loading="loading"
          >
            确定
          </a-button>
          <a-button @click="handleCancel">
            取消
          </a-button>
        </a-space>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  type: {
    type: String,
    default: 'add',
    validator: value => ['add', 'edit'].includes(value)
  },
  device: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:modelValue', 'success'])

const visible = ref(false)
const loading = ref(false)

const form = reactive({
  deviceid: '',
  name: '',
  pwd: '',
  host: '',
  deviceId: '',
  manufacturer: '',
  model: ''
})

watch(
  () => props.modelValue,
  val => {
    visible.value = val
    if (val) {
      resetForm()
      if (props.type === 'edit' && props.device) {
        form.deviceid = props.device.deviceid || ''
        form.name = props.device.name || ''
        form.pwd = ''
        form.host = props.device.host || ''
        form.manufacturer = props.device.manufacturer || ''
        form.model = props.device.model || ''
      }
    }
  },
  { immediate: true }
)

watch(visible, val => {
  emit('update:modelValue', val)
})

const resetForm = () => {
  form.deviceid = ''
  form.name = ''
  form.pwd = ''
  form.host = ''
  form.deviceId = ''
  form.manufacturer = ''
  form.model = ''
}

const handleCancel = () => {
  visible.value = false
}

const handleSubmit = async () => {
  loading.value = true
  try {
    emit('success', { ...form })
  } finally {
    loading.value = false
  }
}
</script>
