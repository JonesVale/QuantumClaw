import { 
  Typography, 
  Stack, 
  OutlinedInput, 
  InputAdornment, 
  Button, 
  InputLabel, 
  FormControl,
  Box,
  Chip,
  Grid,
  Divider,
  CircularProgress,
  Alert
} from '@mui/material';
import { IconWallet, IconCreditCard, IconBrandStripe } from '@tabler/icons-react';
import { useTheme } from '@mui/material/styles';
import SubCard from 'ui-component/cards/SubCard';
import UserCard from 'ui-component/cards/UserCard';

import { API } from 'utils/api';
import React, { useEffect, useState, useCallback } from 'react';
import { showError, showInfo, showSuccess, renderQuota } from 'utils/common';

// ==================== 增强版支付卡片组件 ====================

// 支付方式颜色映射
const PaymentMethodColors = {
  epay: '#1976d2',
  stripe: '#635bff',
  creem: '#22c55e',
  waffo: '#f97316',
};

const TopupCard = () => {
  const theme = useTheme();
  
  // 基础状态
  const [redemptionCode, setRedemptionCode] = useState('');
  const [topUpLink, setTopUpLink] = useState('');
  const [userQuota, setUserQuota] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  // 支付信息状态
  const [paymentInfo, setPaymentInfo] = useState(null);
  const [selectedAmount, setSelectedAmount] = useState(null);
  const [selectedPaymentMethod, setSelectedPaymentMethod] = useState(null);
  const [customAmount, setCustomAmount] = useState('');
  const [isLoadingPaymentInfo, setIsLoadingPaymentInfo] = useState(false);
  const [isProcessingPayment, setIsProcessingPayment] = useState(false);
  const [paymentError, setPaymentError] = useState('');

  // 获取支付配置信息
  const fetchPaymentInfo = useCallback(async () => {
    setIsLoadingPaymentInfo(true);
    try {
      const res = await API.get('/api/user/topup/info');
      const { success, message, data } = res.data;
      if (success) {
        setPaymentInfo(data);
        // 默认选择第一个启用的支付方式
        if (data.enable_stripe_topup) {
          setSelectedPaymentMethod('stripe');
        } else if (data.enable_epay) {
          setSelectedPaymentMethod('epay');
        }
      }
    } catch (err) {
      console.error('获取支付信息失败:', err);
    } finally {
      setIsLoadingPaymentInfo(false);
    }
  }, []);

  // 获取用户配额
  const getUserQuota = async () => {
    try {
      const res = await API.get('/api/user/self');
      const { success, data } = res.data;
      if (success) {
        setUserQuota(data.quota);
      }
    } catch (err) {
      console.error('获取用户配额失败:', err);
    }
  };

  // 兑换码充值
  const topUp = async () => {
    if (redemptionCode === '') {
      showInfo('请输入充值码！');
      return;
    }
    setIsSubmitting(true);
    try {
      const res = await API.post('/api/user/topup', {
        key: redemptionCode
      });
      const { success, message, data } = res.data;
      if (success) {
        showSuccess('充值成功！');
        setUserQuota((quota) => quota + data);
        setRedemptionCode('');
        getUserQuota();
      } else {
        showError(message);
      }
    } catch (err) {
      showError('请求失败');
    } finally {
      setIsSubmitting(false);
    }
  };

  // 处理支付请求
  const handlePayment = async (paymentMethod) => {
    if (!selectedAmount && !customAmount) {
      setPaymentError('请选择或输入充值金额');
      return;
    }
    
    const amount = selectedAmount || parseInt(customAmount, 10);
    if (!amount || amount <= 0) {
      setPaymentError('请输入有效的充值金额');
      return;
    }

    setIsProcessingPayment(true);
    setPaymentError('');
    
    try {
      let res;
      
      switch (paymentMethod) {
        case 'stripe':
          res = await API.post('/api/user/topup/stripe', {
            amount: amount,
            payment_method: 'stripe'
          });
          break;
        case 'creem':
          res = await API.post('/api/user/topup/creem', {
            product_id: amount,
            payment_method: 'creem'
          });
          break;
        case 'waffo':
          res = await API.post('/api/user/topup/waffo', {
            amount: amount,
            pay_method_type: 'default'
          });
          break;
        case 'epay':
        default:
          res = await API.post('/api/user/topup/epay', {
            amount: amount,
            payment_method: 'epay'
          });
          break;
      }
      
      const { success, message, data } = res.data;
      
      if (success) {
        showSuccess('订单创建成功！正在跳转到支付页面...');
        
        // 跳转到支付页面
        if (data.checkout_url) {
          window.open(data.checkout_url, '_blank');
        } else if (data.pay_link) {
          window.open(data.pay_link, '_blank');
        }
        
        // 刷新配额
        setTimeout(() => {
          getUserQuota();
        }, 3000);
      } else {
        setPaymentError(message || '创建订单失败');
        showError(message);
      }
    } catch (err) {
      const errorMsg = err.response?.data?.message || '支付请求失败';
      setPaymentError(errorMsg);
      showError(errorMsg);
    } finally {
      setIsProcessingPayment(false);
    }
  };

  useEffect(() => {
    let status = localStorage.getItem('siteInfo');
    if (status) {
      status = JSON.parse(status);
      if (status.top_up_link) {
        setTopUpLink(status.top_up_link);
      }
    }
    
    // 获取支付配置和用户配额
    Promise.all([fetchPaymentInfo(), getUserQuota()]);
  }, [fetchPaymentInfo]);

  // 判断是否有在线支付功能
  const hasOnlinePayment = paymentInfo && (
    paymentInfo.enable_stripe_topup || 
    paymentInfo.enable_creem_topup || 
    paymentInfo.enable_waffo_topup ||
    paymentInfo.enable_online_topup
  );

  return (
    <UserCard>
      {/* 额度显示 */}
      <Stack 
        direction="row" 
        alignItems="center" 
        justifyContent="center" 
        spacing={2} 
        paddingTop={'20px'}
        paddingBottom={'20px'}
      >
        <IconWallet color={theme.palette.primary.main} />
        <Typography variant="h4">当前额度:</Typography>
        <Typography variant="h4" color="primary">
          {renderQuota(userQuota)}
        </Typography>
      </Stack>

      <Divider sx={{ mb: 2 }} />

      <Grid container spacing={2}>
        {/* 左侧：在线支付 */}
        {hasOnlinePayment && (
          <Grid item xs={12} md={8}>
            <SubCard 
              title="💳 在线充值" 
              secondary={
                <Chip label="安全支付" color="success" size="small" />
              }
            >
              {isLoadingPaymentInfo ? (
                <Stack alignItems="center" spacing={2} py={3}>
                  <CircularProgress size={24} />
                  <Typography color="textSecondary">加载支付信息...</Typography>
                </Stack>
              ) : (
                <>
                  {/* 支付方式选择 */}
                  <Box mb={3}>
                    <Typography variant="subtitle2" gutterBottom>
                      选择支付方式：
                    </Typography>
                    <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                      {paymentInfo?.enable_stripe_topup && (
                        <Chip
                          icon={<IconBrandStripe />}
                          label="Stripe"
                          onClick={() => setSelectedPaymentMethod('stripe')}
                          color={selectedPaymentMethod === 'stripe' ? 'primary' : 'default'}
                          variant={selectedPaymentMethod === 'stripe' ? 'filled' : 'outlined'}
                          sx={{ 
                            borderColor: PaymentMethodColors.stripe,
                            '& .MuiChip-icon': { color: PaymentMethodColors.stripe }
                          }}
                        />
                      )}
                      {paymentInfo?.enable_creem_topup && (
                        <Chip
                          label="🌐 Creem"
                          onClick={() => setSelectedPaymentMethod('creem')}
                          color={selectedPaymentMethod === 'creem' ? 'primary' : 'default'}
                          variant={selectedPaymentMethod === 'creem' ? 'filled' : 'outlined'}
                          sx={{ borderColor: PaymentMethodColors.creem }}
                        />
                      )}
                      {paymentInfo?.enable_waffo_topup && (
                        <Chip
                          label="💰 Waffo"
                          onClick={() => setSelectedPaymentMethod('waffo')}
                          color={selectedPaymentMethod === 'waffo' ? 'primary' : 'default'}
                          variant={selectedPaymentMethod === 'waffo' ? 'filled' : 'outlined'}
                          sx={{ borderColor: PaymentMethodColors.waffo }}
                        />
                      )}
                      {paymentInfo?.enable_online_topup && (
                        <Chip
                          icon={<IconCreditCard />}
                          label="易支付"
                          onClick={() => setSelectedPaymentMethod('epay')}
                          color={selectedPaymentMethod === 'epay' ? 'primary' : 'default'}
                          variant={selectedPaymentMethod === 'epay' ? 'filled' : 'outlined'}
                        />
                      )}
                    </Stack>
                  </Box>

                  {/* 金额选择 */}
                  <Box mb={2}>
                    <Typography variant="subtitle2" gutterBottom>
                      选择充值金额：
                    </Typography>
                    <Grid container spacing={1}>
                      {(paymentInfo?.amount_options || [100, 500, 1000, 5000]).map((amount) => (
                        <Grid item key={amount}>
                          <Button
                            variant={selectedAmount === amount ? 'contained' : 'outlined'}
                            onClick={() => {
                              setSelectedAmount(amount);
                              setCustomAmount('');
                              setPaymentError('');
                            }}
                            sx={{ minWidth: 80 }}
                          >
                            {amount}
                          </Button>
                        </Grid>
                      ))}
                    </Grid>
                  </Box>

                  {/* 自定义金额 */}
                  <FormControl fullWidth sx={{ mb: 2 }}>
                    <InputLabel htmlFor="custom-amount">自定义金额</InputLabel>
                    <OutlinedInput
                      id="custom-amount"
                      type="number"
                      value={customAmount}
                      onChange={(e) => {
                        setCustomAmount(e.target.value);
                        setSelectedAmount(null);
                        setPaymentError('');
                      }}
                      label="自定义金额"
                      placeholder="输入自定义金额"
                    />
                  </FormControl>

                  {/* 错误提示 */}
                  {paymentError && (
                    <Alert severity="error" sx={{ mb: 2 }} onClose={() => setPaymentError('')}>
                      {paymentError}
                    </Alert>
                  )}

                  {/* 支付按钮 */}
                  <Button
                    variant="contained"
                    color="primary"
                    size="large"
                    fullWidth
                    onClick={() => handlePayment(selectedPaymentMethod)}
                    disabled={isProcessingPayment || !selectedPaymentMethod}
                    startIcon={isProcessingPayment ? <CircularProgress size={20} color="inherit" /> : <IconCreditCard />}
                  >
                    {isProcessingPayment ? '处理中...' : '立即充值'}
                  </Button>

                  {/* 支付说明 */}
                  <Box mt={2}>
                    <Typography variant="caption" color="textSecondary">
                      💡 支付成功后，额度将自动添加到您的账户。
                      如有问题，请联系管理员。
                    </Typography>
                  </Box>
                </>
              )}
            </SubCard>
          </Grid>
        )}

        {/* 右侧：兑换码充值 */}
        <Grid item xs={12} md={hasOnlinePayment ? 4 : 12}>
          <SubCard 
            title="🎫 兑换码充值" 
            secondary={
              hasOnlinePayment ? null : (
                <Chip label="推荐" color="primary" size="small" />
              )
            }
          >
            <FormControl fullWidth variant="outlined" sx={{ mb: 2 }}>
              <InputLabel htmlFor="key">兑换码</InputLabel>
              <OutlinedInput
                id="key"
                label="兑换码"
                type="text"
                value={redemptionCode}
                onChange={(e) => setRedemptionCode(e.target.value)}
                placeholder="请输入兑换码"
                endAdornment={
                  <InputAdornment position="end">
                    <Button 
                      variant="contained" 
                      onClick={topUp} 
                      disabled={isSubmitting}
                    >
                      {isSubmitting ? '兑换中...' : '兑换'}
                    </Button>
                  </InputAdornment>
                }
              />
            </FormControl>

            {topUpLink && (
              <Button
                variant="outlined"
                fullWidth
                onClick={() => window.open(topUpLink, '_blank')}
                sx={{ mt: 1 }}
              >
                获取兑换码
              </Button>
            )}
          </SubCard>
        </Grid>
      </Grid>
    </UserCard>
  );
};

export default TopupCard;
