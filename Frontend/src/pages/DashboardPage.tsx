import { useQuery } from '@tanstack/react-query'
import { Card, Col, Row, Statistic, Typography, Alert } from 'antd'
import {
  FileTextOutlined,
  TeamOutlined,
  AuditOutlined,
} from '@ant-design/icons'
import { useMe } from '../hooks/useMe'
import { iamApi } from '../api/iam'

const { Title } = Typography

export default function DashboardPage() {
  const { data: me } = useMe()

  const { data: users } = useQuery({
    queryKey: ['users'],
    queryFn: iamApi.listUsers,
  })

  const { data: departments } = useQuery({
    queryKey: ['departments'],
    queryFn: iamApi.listDepartments,
  })

  return (
    <div>
      <Title level={3}>Добро пожаловать, {me?.login}</Title>
      {me?.is_head && (
        <Alert
          message="Вы являетесь начальником отдела"
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="Сотрудников"
              value={users?.length ?? 0}
              prefix={<TeamOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="Отделов"
              value={departments?.length ?? 0}
              prefix={<AuditOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="Мой отдел"
              value={me?.department_id || '—'}
              prefix={<FileTextOutlined />}
            />
          </Card>
        </Col>
      </Row>
    </div>
  )
}
