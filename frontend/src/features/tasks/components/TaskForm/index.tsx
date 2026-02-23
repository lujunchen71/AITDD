import React from 'react';
import { Modal, Form, Input, Select } from 'antd';
import type { FormProps } from 'antd';

interface TaskFormProps {
  open: boolean;
  moduleId: string;
  initialValues?: {
    id?: string;
    name?: string;
    description?: string;
    prompt?: string;
    status?: string;
    assignee?: string;
    upstreamContractDetail?: string;
    downstreamContractDetail?: string;
  };
  onSubmit: (values: any) => void;
  onCancel: () => void;
  loading?: boolean;
}

const statusOptions = [
  { value: 'ready', label: '就绪' },
  { value: 'in_progress', label: '进行中' },
  { value: 'done', label: '已完成' },
  { value: 'blocked', label: '阻塞' },
];

const TaskForm: React.FC<TaskFormProps> = ({
  open,
  moduleId,
  initialValues,
  onSubmit,
  onCancel,
  loading,
}) => {
  const [form] = Form.useForm();
  const isEdit = !!initialValues?.id;

  const handleSubmit: FormProps['onFinish'] = (values) => {
    onSubmit({
      ...values,
      moduleId,
      version: initialValues?.version || 1,
    });
  };

  return (
    <Modal
      open={open}
      title={isEdit ? '编辑任务' : '新建任务'}
      onOk={() => form.submit()}
      onCancel={onCancel}
      confirmLoading={loading}
      width={700}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          status: 'ready',
          ...initialValues,
        }}
        onFinish={handleSubmit}
      >
        <Form.Item
          name="name"
          label="任务名称"
          rules={[{ required: true, message: '请输入任务名称' }]}
        >
          <Input placeholder="请输入任务名称" />
        </Form.Item>

        <Form.Item
          name="description"
          label="描述"
        >
          <Input.TextArea rows={2} placeholder="请输入任务描述" />
        </Form.Item>

        <Form.Item
          name="prompt"
          label="Prompt"
        >
          <Input.TextArea
            rows={4}
            placeholder="请输入给AI的Prompt指令"
            showCount
            maxLength={10000}
          />
        </Form.Item>

        <Form.Item
          name="status"
          label="状态"
          rules={[{ required: true }]}
        >
          <Select options={statusOptions} />
        </Form.Item>

        <Form.Item
          name="assignee"
          label="分配给"
        >
          <Input placeholder="AI代理标识（如：KiloCode-1）" />
        </Form.Item>

        <Form.Item
          name="upstreamContractDetail"
          label="上游契约详情"
        >
          <Input.TextArea
            rows={3}
            placeholder="上游模块/任务的契约接口详情"
          />
        </Form.Item>

        <Form.Item
          name="downstreamContractDetail"
          label="下游契约详情"
        >
          <Input.TextArea
            rows={3}
            placeholder="下游模块/任务的契约接口详情"
          />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default TaskForm;
