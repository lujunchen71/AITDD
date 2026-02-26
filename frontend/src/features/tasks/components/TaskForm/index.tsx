import React, { useEffect, useState } from 'react';
import { Modal, Form, Input, Select, Button, Space, Card } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import type { FormProps } from 'antd';

interface TestItem {
  name: string;
  description: string;
  precondition: string;
  steps: string[];
  expected: string;
}

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
  const [newStep, setNewStep] = useState<{ [key: number]: string }>({});

  // 解析初始值中的tests字段
  useEffect(() => {
    if (initialValues?.tests) {
      try {
        const parsedTests = JSON.parse(initialValues.tests);
        if (Array.isArray(parsedTests)) {
          setTests(parsedTests);
        } else {
          setTests([]);
        }
      } catch {
        setTests([]);
      }
    } else {
      setTests([]);
    }
    setNewStep({});
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
    setTests([...tests, { name: '', description: '', precondition: '', steps: [], expected: '' }]);
  };

  const removeTest = (index: number) => {
    const newTests = tests.filter((_, i) => i !== index);
    setTests(newTests);
  };

  const updateTest = (index: number, field: keyof TestItem, value: string | string[]) => {
    const newTests = [...tests];
    newTests[index] = { ...newTests[index], [field]: value };
    setTests(newTests);
  };

  const addStep = (testIndex: number) => {
    const step = newStep[testIndex]?.trim();
    if (step) {
      const newTests = [...tests];
      newTests[testIndex] = {
        ...newTests[testIndex],
        steps: [...newTests[testIndex].steps, step],
      };
      setTests(newTests);
      setNewStep({ ...newStep, [testIndex]: '' });
    }
  };

  const removeStep = (testIndex: number, stepIndex: number) => {
    const newTests = [...tests];
    newTests[testIndex] = {
      ...newTests[testIndex],
      steps: newTests[testIndex].steps.filter((_, i) => i !== stepIndex),
    };
    setTests(newTests);
  };

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
          {tests.map((test, index) => (
            <Card
              key={index}
              size="small"
              style={{ marginBottom: 8 }}
              title={`测试用例 ${index + 1}${test.name ? `: ${test.name}` : ''}`}
              extra={
                <Button
                  type="text"
                  danger
                  icon={<DeleteOutlined />}
                  onClick={() => removeTest(index)}
                />
              }
            >
              <Space direction="vertical" style={{ width: '100%' }} size="small">
                <div>
                  <label style={{ fontSize: 12, color: '#666', marginBottom: 4, display: 'block' }}>名称 *</label>
                  <Input
                    placeholder="测试用例名称"
                    value={test.name}
                    onChange={(e) => updateTest(index, 'name', e.target.value)}
                  />
                </div>
                <div>
                  <label style={{ fontSize: 12, color: '#666', marginBottom: 4, display: 'block' }}>描述</label>
                  <Input.TextArea
                    placeholder="测试用例描述"
                    rows={2}
                    value={test.description}
                    onChange={(e) => updateTest(index, 'description', e.target.value)}
                  />
                </div>
                <div>
                  <label style={{ fontSize: 12, color: '#666', marginBottom: 4, display: 'block' }}>前置条件</label>
                  <Input
                    placeholder="执行测试的前置条件"
                    value={test.precondition}
                    onChange={(e) => updateTest(index, 'precondition', e.target.value)}
                  />
                </div>
                <div>
                  <label style={{ fontSize: 12, color: '#666', marginBottom: 4, display: 'block' }}>测试步骤</label>
                  {test.steps.map((step, stepIdx) => (
                    <div key={stepIdx} style={{ display: 'flex', alignItems: 'center', marginBottom: 4 }}>
                      <span style={{ color: '#1890ff', marginRight: 8 }}>{stepIdx + 1}.</span>
                      <Input
                        value={step}
                        onChange={(e) => {
                          const newSteps = [...test.steps];
                          newSteps[stepIdx] = e.target.value;
                          updateTest(index, 'steps', newSteps);
                        }}
                        style={{ flex: 1 }}
                      />
                      <Button
                        type="text"
                        danger
                        size="small"
                        icon={<DeleteOutlined />}
                        onClick={() => removeStep(index, stepIdx)}
                      />
                    </div>
                  ))}
                  <div style={{ display: 'flex', gap: 8 }}>
                    <Input
                      placeholder="添加新步骤"
                      value={newStep[index] || ''}
                      onChange={(e) => setNewStep({ ...newStep, [index]: e.target.value })}
                      onPressEnter={() => addStep(index)}
                    />
                    <Button size="small" onClick={() => addStep(index)}>添加</Button>
                  </div>
                </div>
                <div>
                  <label style={{ fontSize: 12, color: '#666', marginBottom: 4, display: 'block' }}>预期结果</label>
                  <Input.TextArea
                    placeholder="测试的预期结果"
                    rows={2}
                    value={test.expected}
                    onChange={(e) => updateTest(index, 'expected', e.target.value)}
                  />
                </div>
              </Space>
            </Card>
          ))}
          {tests.length === 0 && (
            <div style={{ color: '#999', textAlign: 'center', padding: 16 }}>
              暂无测试用例，点击上方按钮添加
            </div>
          )}
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default TaskForm;
