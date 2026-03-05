import React from 'react';
import { Modal, Form, Input, Select, InputNumber } from 'antd';
import type { FormProps } from 'antd';

interface ModuleFormProps {
  open: boolean;
  initialValues?: {
    id?: string;
    name?: string;
    description?: string;
    prompt?: string;
    status?: string;
    testCoverage?: number;
    parentId?: string;
    version?: number;
  };
  parentId?: string;
  onSubmit: (values: any) => void;
  onCancel: () => void;
  loading?: boolean;
}

const statusOptions = [
  { value: 'designing', label: '设计中' },
  { value: 'developing', label: '开发中' },
  { value: 'testing', label: '测试中' },
  { value: 'done', label: '已完成' },
];

const ModuleForm: React.FC<ModuleFormProps> = ({
  open,
  initialValues,
  parentId,
  onSubmit,
  onCancel,
  loading,
}) => {
  const [form] = Form.useForm();
  const isEdit = !!initialValues?.id;

  const handleSubmit: FormProps['onFinish'] = (values) => {
    onSubmit({
      ...values,
      parentId: parentId || initialValues?.parentId || null,
      version: initialValues?.version || 1,
    });
  };

  return (
    <Modal
      open={open}
      title={isEdit ? '编辑模块' : '新建模块'}
      onOk={() => form.submit()}
      onCancel={onCancel}
      confirmLoading={loading}
      width={600}
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          status: 'designing',
          ...initialValues,
        }}
        onFinish={handleSubmit}
      >
        <Form.Item
          name="name"
          label="模块名称"
          rules={[{ required: true, message: '请输入模块名称' }]}
        >
          <Input placeholder="请输入模块名称" />
        </Form.Item>

        <Form.Item
          name="description"
          label="描述"
        >
          <Input.TextArea rows={3} placeholder="请输入模块描述" />
        </Form.Item>

        <Form.Item
          name="prompt"
          label="Prompt"
        >
          <Input.TextArea
            rows={5}
            placeholder="请输入给AI的Prompt指令"
            showCount
            maxLength={5000}
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
          name="testCoverage"
          label="测试覆盖率 (%)"
        >
          <InputNumber
            min={0}
            max={100}
            style={{ width: '100%' }}
            placeholder="0-100"
          />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default ModuleForm;
