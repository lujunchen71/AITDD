import React, { useEffect, useState } from 'react';
import { Modal, Form, Input, Select, Button, Table } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import type { FormProps } from 'antd';
import { TestItem } from '../../../../types';

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
    tests?: string;
    version?: number;
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
  const [tests, setTests] = useState<TestItem[]>([]);

  // 解析初始值中的tests字段
  useEffect(() => {
    if (initialValues?.tests) {
      try {
        const parsedTests = JSON.parse(initialValues.tests);
        if (Array.isArray(parsedTests)) {
          // 统一转换为 MCP 格式 {target, api}
          const normalizedTests = parsedTests.map((test: any) => ({
            target: test.target || test.name || '',
            api: test.api || test.expected || '',
          }));
          setTests(normalizedTests);
        } else {
          setTests([]);
        }
      } catch {
        setTests([]);
      }
    } else {
      setTests([]);
    }
  }, [initialValues?.tests, open]);

  const handleSubmit: FormProps['onFinish'] = (values) => {
    onSubmit({
      ...values,
      moduleId,
      version: initialValues?.version || 1,
      tests: tests.length > 0 ? JSON.stringify(tests) : undefined,
    });
  };

  const addTest = () => {
    setTests([...tests, { target: '', api: '' }]);
  };

  const removeTest = (index: number) => {
    const newTests = tests.filter((_, i) => i !== index);
    setTests(newTests);
  };

  const updateTest = (index: number, field: keyof TestItem, value: string) => {
    const newTests = [...tests];
    newTests[index] = { ...newTests[index], [field]: value };
    setTests(newTests);
  };

  // 测试用例表格列定义
  const testColumns = [
    {
      title: '#',
      key: 'index',
      width: 40,
      render: (_: any, __: any, index: number) => index + 1,
    },
    {
      title: '测试目标',
      dataIndex: 'target',
      key: 'target',
      render: (value: string, _record: TestItem, index: number) => (
        <Input
          placeholder="描述测试目标，如：验证构造函数正确初始化"
          value={value}
          onChange={(e) => updateTest(index, 'target', e.target.value)}
        />
      ),
    },
    {
      title: '测试API',
      dataIndex: 'api',
      key: 'api',
      width: 250,
      render: (value: string, _record: TestItem, index: number) => (
        <Input
          placeholder="测试函数名，如：test_constructor()"
          value={value}
          onChange={(e) => updateTest(index, 'api', e.target.value)}
          style={{ fontFamily: 'monospace' }}
        />
      ),
    },
    {
      title: '',
      key: 'action',
      width: 40,
      render: (_: any, _record: TestItem, index: number) => (
        <Button
          type="text"
          danger
          size="small"
          icon={<DeleteOutlined />}
          onClick={() => removeTest(index)}
        />
      ),
    },
  ];

  return (
    <Modal
      open={open}
      title={isEdit ? '编辑任务' : '新建任务'}
      onOk={() => form.submit()}
      onCancel={onCancel}
      confirmLoading={loading}
      width={800}
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

        <Form.Item label="测试用例">
          <div style={{ marginBottom: 8 }}>
            <Button type="dashed" onClick={addTest} icon={<PlusOutlined />} block>
              添加测试用例
            </Button>
          </div>
          
          {tests.length > 0 ? (
            <Table
              dataSource={tests}
              columns={testColumns}
              pagination={false}
              size="small"
              rowKey={(_, index) => `test-${index}`}
            />
          ) : (
            <div style={{ color: '#999', textAlign: 'center', padding: 16, border: '1px dashed #d9d9d9', borderRadius: 4 }}>
              暂无测试用例，点击上方按钮添加
            </div>
          )}
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default TaskForm;
